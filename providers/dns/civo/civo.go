// Package civo implements a DNS provider for solving the DNS-01 challenge using CIVO.
package civo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/civo/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "CIVO_"

	EnvAPIToken = envNamespace + "TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const (
	minTTL                    = 600
	defaultPollingInterval    = 30 * time.Second
	defaultPropagationTimeout = 300 * time.Second
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, defaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for CIVO.
// Credentials must be passed in the environment variables: API_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("civo: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for CIVO.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("civo: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("civo: credentials missing")
	}

	if config.TTL < minTTL {
		config.TTL = minTTL
	}

	// Create a Civo client - DNS is region independent, we can use any region
	client, err := internal.NewClient(
		clientdebug.Wrap(
			internal.OAuthStaticAccessToken(config.HTTPClient, config.Token),
		),
		"LON1")
	if err != nil {
		return nil, fmt.Errorf("civo: %w", err)
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("civo: could not find zone for domain %q: %w", domain, err)
	}

	zone := dns01.UnFqdn(authZone)

	domainID, err := d.getDomainIDByName(ctx, zone)
	if err != nil {
		return fmt.Errorf("civo: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("civo: %w", err)
	}

	_, err = d.client.CreateDNSRecord(ctx, domainID, internal.Record{
		Name:  subDomain,
		Value: info.Value,
		Type:  "TXT",
		TTL:   d.config.TTL,
	})
	if err != nil {
		return fmt.Errorf("civo: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("civo: could not find zone for domain %q: %w", domain, err)
	}

	zone := dns01.UnFqdn(authZone)

	domainID, err := d.getDomainIDByName(ctx, zone)
	if err != nil {
		return fmt.Errorf("civo: %w", err)
	}

	dnsRecords, err := d.client.ListDNSRecords(ctx, domainID)
	if err != nil {
		return fmt.Errorf("civo: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("civo: %w", err)
	}

	var dnsRecord internal.Record

	for _, entry := range dnsRecords {
		if entry.Name == subDomain && entry.Value == info.Value {
			dnsRecord = entry
			break
		}
	}

	err = d.client.DeleteDNSRecord(ctx, dnsRecord)
	if err != nil {
		return fmt.Errorf("civo: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getDomainIDByName(ctx context.Context, domain string) (string, error) {
	domains, err := d.client.ListDomains(ctx)
	if err != nil {
		return "", fmt.Errorf("list domains: %w", err)
	}

	for _, d := range domains {
		if d.Name == domain {
			return d.ID, nil
		}
	}

	return "", fmt.Errorf("domain %q not found", domain)
}
