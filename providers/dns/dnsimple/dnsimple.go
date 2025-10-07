// Package dnsimple implements a DNS provider for solving the DNS-01 challenge using dnsimple DNS.
package dnsimple

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dnsimple/dnsimple-go/v4/dnsimple"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	"golang.org/x/oauth2"
)

// Environment variables names.
const (
	envNamespace = "DNSIMPLE_"

	EnvOAuthToken = envNamespace + "OAUTH_TOKEN"
	EnvBaseURL    = envNamespace + "BASE_URL"
	EnvDebug      = envNamespace + "DEBUG"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Debug              bool
	AccessToken        string
	BaseURL            string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		Debug:              env.GetOrDefaultBool(EnvDebug, false),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *dnsimple.Client
}

// NewDNSProvider returns a DNSProvider instance configured for dnsimple.
// Credentials must be passed in the environment variable: DNSIMPLE_OAUTH_TOKEN.
//
// See: https://developer.dnsimple.com/v2/#authentication
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.AccessToken = env.GetOrFile(EnvOAuthToken)
	config.BaseURL = env.GetOrFile(EnvBaseURL)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DNSimple.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnsimple: the configuration of the DNS provider is nil")
	}

	if config.AccessToken == "" {
		return nil, errors.New("dnsimple: OAuth token is missing")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.AccessToken})
	client := dnsimple.NewClient(oauth2.NewClient(context.Background(), ts))
	client.SetUserAgent(useragent.Get())

	if config.BaseURL != "" {
		client.BaseURL = config.BaseURL
	}

	client.Debug = config.Debug

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneName, err := d.getHostedZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dnsimple: %w", err)
	}

	accountID, err := d.getAccountID(ctx)
	if err != nil {
		return fmt.Errorf("dnsimple: %w", err)
	}

	recordAttributes, err := newTxtRecord(zoneName, info.EffectiveFQDN, info.Value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("dnsimple: %w", err)
	}

	_, err = d.client.Zones.CreateRecord(ctx, accountID, zoneName, recordAttributes)
	if err != nil {
		return fmt.Errorf("dnsimple: API call failed: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	records, err := d.findTxtRecords(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dnsimple: %w", err)
	}

	accountID, err := d.getAccountID(ctx)
	if err != nil {
		return fmt.Errorf("dnsimple: %w", err)
	}

	var lastErr error
	for _, rec := range records {
		_, err := d.client.Zones.DeleteRecord(ctx, accountID, rec.ZoneID, rec.ID)
		if err != nil {
			lastErr = fmt.Errorf("dnsimple: %w", err)
		}
	}

	return lastErr
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getHostedZone(ctx context.Context, domain string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return "", fmt.Errorf("could not find zone for FQDN %q: %w", domain, err)
	}

	accountID, err := d.getAccountID(ctx)
	if err != nil {
		return "", err
	}

	hostedZone, err := d.client.Zones.GetZone(ctx, accountID, dns01.UnFqdn(authZone))
	if err != nil {
		return "", fmt.Errorf("get zone: %w", err)
	}

	if hostedZone == nil || hostedZone.Data == nil || hostedZone.Data.ID == 0 {
		return "", fmt.Errorf("zone %s not found in DNSimple for domain %s", authZone, domain)
	}

	return hostedZone.Data.Name, nil
}

func (d *DNSProvider) findTxtRecords(ctx context.Context, fqdn string) ([]dnsimple.ZoneRecord, error) {
	zoneName, err := d.getHostedZone(ctx, fqdn)
	if err != nil {
		return nil, err
	}

	accountID, err := d.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, zoneName)
	if err != nil {
		return nil, err
	}

	result, err := d.client.Zones.ListRecords(ctx, accountID, zoneName, &dnsimple.ZoneRecordListOptions{Name: &subDomain, Type: dnsimple.String("TXT"), ListOptions: dnsimple.ListOptions{}})
	if err != nil {
		return nil, fmt.Errorf("API call has failed: %w", err)
	}

	return result.Data, nil
}

func newTxtRecord(zoneName, fqdn, value string, ttl int) (dnsimple.ZoneRecordAttributes, error) {
	subDomain, err := dns01.ExtractSubDomain(fqdn, zoneName)
	if err != nil {
		return dnsimple.ZoneRecordAttributes{}, err
	}

	return dnsimple.ZoneRecordAttributes{
		Type:    "TXT",
		Name:    &subDomain,
		Content: value,
		TTL:     ttl,
	}, nil
}

func (d *DNSProvider) getAccountID(ctx context.Context) (string, error) {
	whoamiResponse, err := d.client.Identity.Whoami(ctx)
	if err != nil {
		return "", err
	}

	if whoamiResponse.Data.Account == nil {
		return "", errors.New("user tokens are not supported, please use an account token")
	}

	return strconv.FormatInt(whoamiResponse.Data.Account.ID, 10), nil
}
