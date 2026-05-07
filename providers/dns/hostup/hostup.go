// Package hostup implements a DNS provider for solving the DNS-01 challenge using HostUp.
package hostup

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/hostup/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "HOSTUP_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

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

	zoneIDs     map[string]string
	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for HostUp.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("hostup: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for HostUp.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("hostup: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("hostup: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		zoneIDs:   make(map[string]string),
		recordIDs: make(map[string]string),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hostup: could not find zone for domain %q: %w", domain, err)
	}

	zones, err := d.client.ListZones(ctx, dns01.UnFqdn(authZone), 0, "")
	if err != nil {
		return fmt.Errorf("hostup: list zones: %w", err)
	}

	if len(zones) != 1 {
		return fmt.Errorf("hostup: no zone or ambiguous name (count: %d) for %s", len(zones), dns01.UnFqdn(authZone))
	}

	zone := zones[0]

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("hostup: %w", err)
	}

	record := internal.Record{
		Type:  "TXT",
		Name:  subDomain,
		Value: info.Value,
		TTL:   d.config.TTL,
	}

	result, err := d.client.CreateRecord(ctx, zone.ID, record)
	if err != nil {
		return fmt.Errorf("hostup: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.zoneIDs[token] = zone.ID
	d.recordIDs[token] = result.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, recordOK := d.recordIDs[token]
	zoneID, zoneOK := d.zoneIDs[token]
	d.recordIDsMu.Unlock()

	if !zoneOK {
		return fmt.Errorf("hostup: unknown zone ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	if !recordOK {
		return fmt.Errorf("hostup: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecord(ctx, zoneID, recordID)
	if err != nil {
		return fmt.Errorf("hostup: delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
