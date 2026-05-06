// Package hostup implements a DNS provider for solving the DNS-01 challenge using HostUp.
package hostup

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/hostup/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
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
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

type recordRef struct {
	ZoneID   string
	RecordID string
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	records   map[string]recordRef
	recordsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for HostUp.
// Credentials must be passed in the environment variable: HOSTUP_API_KEY.
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

	if config.APIKey == "" {
		return nil, errors.New("hostup: missing credentials")
	}

	client := internal.NewClient(
		clientdebug.Wrap(
			internal.OAuthStaticAccessToken(config.HTTPClient, config.APIKey),
		),
	)

	return &DNSProvider{
		config:  config,
		client:  client,
		records: map[string]recordRef{},
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hostup: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	zone, err := d.findZone(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("hostup: could not find zone for domain %q (%s): %w", domain, authZone, err)
	}

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

	newRecord, err := d.client.AddRecord(ctx, zone.DomainID, record)
	if err != nil {
		return fmt.Errorf("hostup: %w", err)
	}

	d.recordsMu.Lock()
	d.records[token] = recordRef{ZoneID: zone.DomainID, RecordID: newRecord.ID}
	d.recordsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordsMu.Lock()
	ref, ok := d.records[token]
	d.recordsMu.Unlock()

	if !ok {
		return fmt.Errorf("hostup: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecord(context.Background(), ref.ZoneID, ref.RecordID)
	if err != nil {
		return fmt.Errorf("hostup: %w", err)
	}

	d.recordsMu.Lock()
	delete(d.records, token)
	d.recordsMu.Unlock()

	return nil
}

func (d *DNSProvider) findZone(ctx context.Context, domain string) (*internal.Zone, error) {
	zones, err := d.client.GetZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("get zones: %w", err)
	}

	for _, zone := range zones {
		if zone.Domain == domain {
			return &zone, nil
		}
	}

	return nil, fmt.Errorf("zone %q not found", domain)
}
