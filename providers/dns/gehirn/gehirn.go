// Package gehirn implements a DNS provider for solving the DNS-01 challenge using Gehirn.
package gehirn

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
	"github.com/go-acme/lego/v5/providers/dns/gehirn/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "GEHIRN_"

	EnvTokenID     = envNamespace + "TOKEN_ID"
	EnvTokenSecret = envNamespace + "TOKEN_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	TokenID     string
	TokenSecret string

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

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Gehirn.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvTokenID, EnvTokenSecret)
	if err != nil {
		return nil, fmt.Errorf("gehirn: %w", err)
	}

	config := NewDefaultConfig()
	config.TokenID = values[EnvTokenID]
	config.TokenSecret = values[EnvTokenSecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Gehirn.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gehirn: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.TokenID, config.TokenSecret)
	if err != nil {
		return nil, fmt.Errorf("gehirn: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]string),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("gehirn: could not find zone for domain %q: %w", domain, err)
	}

	zone, err := d.findZone(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("gehirn: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("gehirn: %w", err)
	}

	record := internal.Record{
		Name: subDomain,
		Type: "TXT",
		TTL:  d.config.TTL,
		Records: []internal.RecordTXT{
			{Data: info.Value},
		},
	}

	newRecord, err := d.client.CreateRecord(ctx, zone.ID, zone.CurrentVersionID, record)
	if err != nil {
		return fmt.Errorf("gehirn: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("gehirn: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("gehirn: could not find zone for domain %q: %w", domain, err)
	}

	zone, err := d.findZone(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("gehirn: %w", err)
	}

	_, err = d.client.DeleteRecord(ctx, zone.ID, zone.CurrentVersionID, recordID)
	if err != nil {
		return fmt.Errorf("gehirn: delete record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, domain string) (*internal.Zone, error) {
	zones, err := d.client.ListZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("list zones: %w", err)
	}

	for _, zone := range zones {
		if zone.Name == domain {
			return &zone, nil
		}
	}

	return nil, fmt.Errorf("zone not found: %s", domain)
}
