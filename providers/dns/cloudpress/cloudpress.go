// Package cloudpress implements a DNS provider for solving the DNS-01 challenge using CloudPress.
package cloudpress

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
	"github.com/go-acme/lego/v5/providers/dns/cloudpress/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "CLOUDPRESS_"

	EnvBaseURL   = envNamespace + "BASE_URL"
	EnvAPIKey    = envNamespace + "API_KEY"
	EnvAccountID = envNamespace + "ACCOUNT_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const maxPerPage = 100

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL   string
	APIKey    string
	AccountID string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 5*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
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
	zoneIDs     map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for CloudPress.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvBaseURL, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("cloudpress: %w", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values[EnvBaseURL]
	config.APIKey = values[EnvAPIKey]
	config.AccountID = env.GetOrFile(EnvAccountID)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for CloudPress.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("cloudpress: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.BaseURL, config.APIKey, config.AccountID)
	if err != nil {
		return nil, fmt.Errorf("cloudpress: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]string),
		zoneIDs:   make(map[string]string),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("cloudpress: could not find zone for domain %q: %w", domain, err)
	}

	zoneID, err := d.findZoneID(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("cloudpress: %w", err)
	}

	record := internal.Record{
		Type:    internal.RecordTypeTXT,
		Name:    dns01.UnFqdn(info.EffectiveFQDN),
		Value:   info.Value,
		TTL:     d.config.TTL,
		Comment: "Created by go-acme/lego",
	}

	newRecord, err := d.client.CreateRecord(ctx, zoneID, record)
	if err != nil {
		return fmt.Errorf("cloudpress: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.zoneIDs[token] = zoneID
	d.recordIDs[token] = newRecord.ID
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

	if !recordOK {
		return fmt.Errorf("cloudpress: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	if !zoneOK {
		return fmt.Errorf("cloudpress: unknown zone ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecord(ctx, zoneID, recordID)
	if err != nil {
		return fmt.Errorf("cloudpress: delete record: %w", err)
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

func (d *DNSProvider) findZoneID(ctx context.Context, authZone string) (string, error) {
	pager := &internal.Pager{Page: 1, PerPage: maxPerPage}

	for {
		zones, err := d.client.ListZones(ctx, pager)
		if err != nil {
			return "", fmt.Errorf("list zones: %w", err)
		}

		for _, zone := range zones {
			if zone.Name == authZone {
				return zone.ID, nil
			}
		}

		if len(zones) < maxPerPage {
			break
		}

		pager.Page++
	}

	return "", fmt.Errorf("could not find the zone for domain %q", authZone)
}
