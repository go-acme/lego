// Package legacy implements a DNS provider for solving the DNS-01 challenge using Hetzner DNS.
package legacy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/hetzner/internal/legacy/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "HETZNER_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const minTTL = 60

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
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
}

// NewDNSProvider returns a DNSProvider instance configured for hetzner.
// Credentials must be passed in the environment variable: HETZNER_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("hetzner (legacy): %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for hetzner.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("hetzner (legacy): the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("hetzner (legacy): credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("hetzner (legacy): invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := internal.NewClient(config.APIKey)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hetzner (legacy): could not find zone for domain %q: %w", domain, err)
	}

	zone := dns01.UnFqdn(authZone)

	ctx := context.Background()

	zoneID, err := d.client.GetZoneID(ctx, zone)
	if err != nil {
		return fmt.Errorf("hetzner (legacy): %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("hetzner (legacy): %w", err)
	}

	record := internal.DNSRecord{
		Type:   "TXT",
		Name:   subDomain,
		Value:  info.Value,
		TTL:    d.config.TTL,
		ZoneID: zoneID,
	}

	if err := d.client.CreateRecord(ctx, record); err != nil {
		return fmt.Errorf("hetzner (legacy): failed to add TXT record: fqdn=%s, zoneID=%s: %w", info.EffectiveFQDN, zoneID, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hetzner (legacy): could not find zone for domain %q: %w", domain, err)
	}

	zone := dns01.UnFqdn(authZone)

	ctx := context.Background()

	zoneID, err := d.client.GetZoneID(ctx, zone)
	if err != nil {
		return fmt.Errorf("hetzner (legacy): %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("hetzner (legacy): %w", err)
	}

	record, err := d.client.GetTxtRecord(ctx, subDomain, info.Value, zoneID)
	if err != nil {
		return fmt.Errorf("hetzner (legacy): %w", err)
	}

	if err := d.client.DeleteRecord(ctx, record.ID); err != nil {
		return fmt.Errorf("hetzner (legacy): failed to delete TXT record: id=%s, name=%s: %w", record.ID, record.Name, err)
	}

	return nil
}
