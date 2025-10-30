// Package keyhelp implements a DNS provider for solving the DNS-01 challenge using KeyHelp.
package keyhelp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/keyhelp/internal"
)

// Environment variables names.
const (
	envNamespace = "KEYHELP_"

	EnvBaseURL = envNamespace + "BASE_URL"
	EnvAPIKey  = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL string
	APIKey  string

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

	domainIDs   map[string]int
	domainIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for KeyHelp.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvBaseURL, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("keyhelp: %w", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values[EnvBaseURL]
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for KeyHelp.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("keyhelp: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.BaseURL, config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("keyhelp: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		domainIDs: make(map[string]int),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("keyhelp: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	domainInfo, err := d.findDomain(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("keyhelp: %w", err)
	}

	domainRecords, err := d.client.ListDomainRecords(ctx, domainInfo.ID)
	if err != nil {
		return fmt.Errorf("keyhelp: list domain records: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("keyhelp: %w", err)
	}

	records := domainRecords.Records.Other
	records = append(records, internal.Record{
		Host:  subDomain,
		TTL:   d.config.TTL,
		Type:  "TXT",
		Value: info.Value,
	})

	req := internal.DomainRecords{
		DkimRecord: domainRecords.DkimRecord,
		Records: &internal.Records{
			Soa:   domainRecords.Records.Soa,
			Other: records,
		},
	}

	_, err = d.client.UpdateDomainRecords(ctx, domainInfo.ID, req)
	if err != nil {
		return fmt.Errorf("keyhelp: update domain records (add): %w", err)
	}

	d.domainIDsMu.Lock()
	d.domainIDs[token] = domainInfo.ID
	d.domainIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	// get the domain's unique ID from when we created it
	d.domainIDsMu.Lock()
	domainID, ok := d.domainIDs[token]
	d.domainIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("keyhelp: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	domainRecords, err := d.client.ListDomainRecords(ctx, domainID)
	if err != nil {
		return fmt.Errorf("keyhelp: list domain records: %w", err)
	}

	var records []internal.Record

	for _, record := range domainRecords.Records.Other {
		if record.Type == "TXT" && record.Value == info.Value {
			continue
		}

		records = append(records, record)
	}

	req := internal.DomainRecords{
		DkimRecord: domainRecords.DkimRecord,
		Records: &internal.Records{
			Soa:   domainRecords.Records.Soa,
			Other: records,
		},
	}

	_, err = d.client.UpdateDomainRecords(ctx, domainID, req)
	if err != nil {
		return fmt.Errorf("keyhelp: update domain records (delete): %w", err)
	}

	// Delete domain ID from map
	d.domainIDsMu.Lock()
	delete(d.domainIDs, token)
	d.domainIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findDomain(ctx context.Context, zone string) (internal.Domain, error) {
	domains, err := d.client.ListDomains(ctx)
	if err != nil {
		return internal.Domain{}, fmt.Errorf("list domains: %w", err)
	}

	for _, domain := range domains {
		if domain.DomainUTF8 == zone || domain.Domain == zone {
			return domain, nil
		}
	}

	return internal.Domain{}, fmt.Errorf("domain not found: %s", zone)
}
