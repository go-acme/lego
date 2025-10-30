// Package octenium implements a DNS provider for solving the DNS-01 challenge using Octenium.
package octenium

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/octenium/internal"
	"github.com/hashicorp/go-retryablehttp"
)

// Environment variables names.
const (
	envNamespace = "OCTENIUM_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	domainIDs   map[string]string
	domainIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Octenium.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("octenium: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Octenium.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("octenium: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("octenium: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.HTTPClient = client.HTTPClient
	retryClient.Logger = log.Logger

	client.HTTPClient = clientdebug.Wrap(retryClient.StandardClient())

	return &DNSProvider{
		config:    config,
		client:    client,
		domainIDs: make(map[string]string),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("octenium: could not find zone for domain '%s': %w", domain, err)
	}

	domainID, err := d.getDomainID(ctx, authZone)
	if err != nil {
		return fmt.Errorf("octenium: get domain ID: %w", err)
	}

	d.domainIDsMu.Lock()
	d.domainIDs[token] = domainID
	d.domainIDsMu.Unlock()

	record := internal.Record{
		Type:  "TXT",
		Name:  info.EffectiveFQDN,
		TTL:   d.config.TTL,
		Value: info.Value,
	}

	_, err = d.client.AddDNSRecord(ctx, domainID, record)
	if err != nil {
		return fmt.Errorf("octenium: add record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.domainIDsMu.Lock()
	domainID, ok := d.domainIDs[token]
	d.domainIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("octenium: unknown domain ID for '%s'", info.EffectiveFQDN)
	}

	records, err := d.client.ListDNSRecords(ctx, domainID, "TXT")
	if err != nil {
		return fmt.Errorf("octenium: list records: %w", err)
	}

	for _, record := range records {
		if record.Type != "TXT" || record.Name != info.EffectiveFQDN || record.Value != info.Value {
			continue
		}

		_, err = d.client.DeleteDNSRecord(ctx, domainID, record.ID)
		if err != nil {
			return fmt.Errorf("octenium: delete record: %w", err)
		}

		break
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getDomainID(ctx context.Context, authZone string) (string, error) {
	domains, err := d.client.ListDomains(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return "", fmt.Errorf("list domains: %w", err)
	}

	if len(domains) == 0 {
		return "", errors.New("domain not found")
	}

	if len(domains) > 1 {
		return "", errors.New("multiple domains found")
	}

	for id := range domains {
		return id, nil
	}

	return "", errors.New("domain ID not found")
}
