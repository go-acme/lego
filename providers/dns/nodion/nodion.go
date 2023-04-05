// Package nodion implements a DNS provider for solving the DNS-01 challenge using Nodion DNS.
package nodion

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/nrdcg/nodion"
)

// Environment variables names.
const (
	envNamespace = "NODION_"

	EnvAPIToken = envNamespace + "API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIToken string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *nodion.Client

	zoneIDs   map[string]string
	zoneIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Nodion.
// Credentials must be passed in the environment variable: NODION_API_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("nodion: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Nodion.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("nodion: the configuration of the DNS provider is nil")
	}

	if config.APIToken == "" {
		return nil, errors.New("nodion: incomplete credentials, missing API token")
	}

	client, err := nodion.NewClient(config.APIToken)
	if err != nil {
		return nil, err
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config:  config,
		client:  client,
		zoneIDs: map[string]string{},
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
		return fmt.Errorf("nodion: could not find zone for domain %q and fqdn %q : %w", domain, info.EffectiveFQDN, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("nodion: %w", err)
	}

	ctx := context.Background()

	zones, err := d.client.GetZones(ctx, &nodion.ZonesFilter{Name: dns01.UnFqdn(authZone)})
	if err != nil {
		return fmt.Errorf("nodion: %w", err)
	}

	if len(zones) == 0 {
		return fmt.Errorf("nodion: zone not found: %s", authZone)
	}

	if len(zones) > 1 {
		return fmt.Errorf("nodion: too many possible zones for the domain %s: %v", authZone, zones)
	}

	zoneID := zones[0].ID

	record := nodion.Record{
		RecordType: nodion.TypeTXT,
		Name:       subDomain,
		Content:    info.Value,
		TTL:        d.config.TTL,
	}

	_, err = d.client.CreateRecord(ctx, zoneID, record)
	if err != nil {
		return fmt.Errorf("nodion: failed to create TXT records [domain: %s, sub domain: %s]: %w",
			dns01.UnFqdn(authZone), subDomain, err)
	}

	d.zoneIDsMu.Lock()
	d.zoneIDs[token] = zoneID
	d.zoneIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("nodion: could not find zone for domain %q and fqdn %q : %w", domain, info.EffectiveFQDN, err)
	}

	d.zoneIDsMu.Lock()
	zoneID, ok := d.zoneIDs[token]
	d.zoneIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("nodion: unknown zone ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("nodion: %w", err)
	}

	ctx := context.Background()

	filter := &nodion.RecordsFilter{
		Name:       subDomain,
		RecordType: nodion.TypeTXT,
		Content:    info.Value,
	}

	records, err := d.client.GetRecords(ctx, zoneID, filter)
	if err != nil {
		return fmt.Errorf("nodion: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("nodion: record not found: %s", authZone)
	}

	if len(records) > 1 {
		return fmt.Errorf("nodion: too many possible records for the domain %s: %v", info.EffectiveFQDN, records)
	}

	_, err = d.client.DeleteRecord(ctx, zoneID, records[0].ID)
	if err != nil {
		return fmt.Errorf("regru: failed to remove TXT records [domain: %s]: %w", dns01.UnFqdn(authZone), err)
	}

	return nil
}
