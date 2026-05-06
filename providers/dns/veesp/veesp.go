// Package veesp implements a DNS provider for solving the DNS-01 challenge using Veesp.
package veesp

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
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/veesp/internal"
)

// Environment variables names.
const (
	envNamespace = "VEESP_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username string
	Password string

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

	Zones   map[string]*internal.Zone
	ZonesMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Veesp.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("veesp: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Veesp.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("veesp: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("veesp: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config: config,
		client: client,
		Zones:  make(map[string]*internal.Zone),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("veesp: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("veesp: %w", err)
	}

	zone, err := d.findZone(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("veesp: %w", err)
	}

	record := internal.Record{
		Name:    subDomain,
		TTL:     d.config.TTL,
		Type:    "TXT",
		Content: info.Value,
	}

	err = d.client.AddRecord(ctx, zone.ServiceID, zone.DomainID, record)
	if err != nil {
		return fmt.Errorf("veesp: add record: %w", err)
	}

	d.ZonesMu.Lock()
	d.Zones[token] = zone
	d.ZonesMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.ZonesMu.Lock()
	zone, ok := d.Zones[token]
	d.ZonesMu.Unlock()

	if !ok {
		return fmt.Errorf("veesp: unknown zone for '%s' '%s'", info.EffectiveFQDN, token)
	}

	record, err := d.findRecord(ctx, zone, info)
	if err != nil {
		return fmt.Errorf("veesp: %w", err)
	}

	err = d.client.RemoveRecord(ctx, zone.ServiceID, zone.DomainID, record.ID)
	if err != nil {
		return fmt.Errorf("veesp: remove record: %w", err)
	}

	d.ZonesMu.Lock()
	delete(d.Zones, token)
	d.ZonesMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, authZone string) (*internal.Zone, error) {
	zones, err := d.client.ListZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("list zones: %w", err)
	}

	for _, zone := range zones {
		if zone.Name == authZone {
			return &zone, nil
		}
	}

	return nil, errors.New("zone not found")
}

func (d *DNSProvider) findRecord(ctx context.Context, zone *internal.Zone, info dns01.ChallengeInfo) (*internal.Record, error) {
	records, err := d.client.GetRecords(ctx, zone.ServiceID, zone.DomainID)
	if err != nil {
		return nil, fmt.Errorf("get records: %w", err)
	}

	for _, record := range records {
		if record.Type == "TXT" && record.Content == info.Value {
			return &record, nil
		}
	}

	return nil, errors.New("record not found")
}
