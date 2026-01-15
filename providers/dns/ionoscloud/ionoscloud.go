// Package ionoscloud implements a DNS provider for solving the DNS-01 challenge using Ionos Cloud.
package ionoscloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge/dnsnew"
	"github.com/go-acme/lego/v5/platform/config/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/ionoscloud/internal"
)

// Environment variables names.
const (
	envNamespace = "IONOSCLOUD_"

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
		TTL:                env.GetOrDefaultInt(EnvTTL, dnsnew.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dnsnew.DefaultPollingInterval),
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

// NewDNSProvider returns a DNSProvider instance configured for Ionos Cloud.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("ionoscloud: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Ionos Cloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ionoscloud: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIToken)
	if err != nil {
		return nil, fmt.Errorf("ionoscloud: %w", err)
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
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dnsnew.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dnsnew.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ionoscloud: could not find zone for domain %q: %w", domain, err)
	}

	zones, err := d.client.RetrieveZones(ctx, dnsnew.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("ionoscloud: retrieve zones: %w", err)
	}

	if len(zones) != 1 {
		return fmt.Errorf("ionoscloud: zone ID not found for domain %q", domain)
	}

	subDomain, err := dnsnew.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("ionoscloud: %w", err)
	}

	zoneID := zones[0].ID

	request := internal.RecordProperties{
		Name:    subDomain,
		Type:    "TXT",
		Content: info.Value,
		TTL:     d.config.TTL,
	}

	record, err := d.client.CreateRecord(ctx, zoneID, request)
	if err != nil {
		return fmt.Errorf("ionoscloud: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.zoneIDs[token] = zoneID
	d.recordIDs[token] = record.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dnsnew.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	zoneID, ok := d.zoneIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("ionoscloud: unknown zone ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("ionoscloud: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecord(ctx, zoneID, recordID)
	if err != nil {
		return fmt.Errorf("ionoscloud: delete record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.zoneIDs, token)
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
