// Package dnsimple implements a DNS provider for solving the DNS-01 challenge using dnsimple DNS.
package dnsimple

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
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
	client.SetUserAgent("go-acme/lego")

	if config.BaseURL != "" {
		client.BaseURL = config.BaseURL
	}

	client.Debug = config.Debug

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zoneName, err := d.getHostedZone(fqdn)
	if err != nil {
		return fmt.Errorf("dnsimple: %w", err)
	}

	accountID, err := d.getAccountID()
	if err != nil {
		return fmt.Errorf("dnsimple: %w", err)
	}

	recordAttributes := newTxtRecord(zoneName, fqdn, value, d.config.TTL)
	_, err = d.client.Zones.CreateRecord(context.Background(), accountID, zoneName, recordAttributes)
	if err != nil {
		return fmt.Errorf("dnsimple: API call failed: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	records, err := d.findTxtRecords(fqdn)
	if err != nil {
		return fmt.Errorf("dnsimple: %w", err)
	}

	accountID, err := d.getAccountID()
	if err != nil {
		return fmt.Errorf("dnsimple: %w", err)
	}

	var lastErr error
	for _, rec := range records {
		_, err := d.client.Zones.DeleteRecord(context.Background(), accountID, rec.ZoneID, rec.ID)
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

func (d *DNSProvider) getHostedZone(domain string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return "", err
	}

	accountID, err := d.getAccountID()
	if err != nil {
		return "", err
	}

	zoneName := dns01.UnFqdn(authZone)

	zones, err := d.client.Zones.ListZones(context.Background(), accountID, &dnsimple.ZoneListOptions{NameLike: &zoneName})
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	var hostedZone dnsimple.Zone
	for _, zone := range zones.Data {
		if zone.Name == zoneName {
			hostedZone = zone
		}
	}

	if hostedZone.ID == 0 {
		return "", fmt.Errorf("zone %s not found in DNSimple for domain %s", authZone, domain)
	}

	return hostedZone.Name, nil
}

func (d *DNSProvider) findTxtRecords(fqdn string) ([]dnsimple.ZoneRecord, error) {
	zoneName, err := d.getHostedZone(fqdn)
	if err != nil {
		return nil, err
	}

	accountID, err := d.getAccountID()
	if err != nil {
		return nil, err
	}

	recordName := extractRecordName(fqdn, zoneName)

	result, err := d.client.Zones.ListRecords(context.Background(), accountID, zoneName, &dnsimple.ZoneRecordListOptions{Name: &recordName, Type: dnsimple.String("TXT"), ListOptions: dnsimple.ListOptions{}})
	if err != nil {
		return nil, fmt.Errorf("API call has failed: %w", err)
	}

	return result.Data, nil
}

func newTxtRecord(zoneName, fqdn, value string, ttl int) dnsimple.ZoneRecordAttributes {
	name := extractRecordName(fqdn, zoneName)

	return dnsimple.ZoneRecordAttributes{
		Type:    "TXT",
		Name:    &name,
		Content: value,
		TTL:     ttl,
	}
}

func extractRecordName(fqdn, zone string) string {
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+zone); idx != -1 {
		return name[:idx]
	}
	return name
}

func (d *DNSProvider) getAccountID() (string, error) {
	whoamiResponse, err := d.client.Identity.Whoami(context.Background())
	if err != nil {
		return "", err
	}

	if whoamiResponse.Data.Account == nil {
		return "", errors.New("user tokens are not supported, please use an account token")
	}

	return strconv.FormatInt(whoamiResponse.Data.Account.ID, 10), nil
}
