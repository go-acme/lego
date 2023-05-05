// Package derak implements a DNS provider for solving the DNS-01 challenge using Derak Cloud.
package derak

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/derak/internal"
	"github.com/miekg/dns"
)

// Environment variables names.
const (
	envNamespace = "DERAK_"

	EnvAPIKey    = envNamespace + "API_KEY"
	EnvWebsiteID = envNamespace + "WEBSITE_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	WebsiteID          string
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
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 5*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for Derak Cloud.
// Credentials must be passed in the environment variable: DERAK_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("derak: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.WebsiteID = env.GetOrDefaultString(EnvWebsiteID, "")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Derak Cloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("derak: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("derak: missing credentials")
	}

	client := internal.NewClient(config.APIKey)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]string),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("derak: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	recordName, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("derak: %w", err)
	}

	zoneID, err := d.getZoneID(ctx, info)
	if err != nil {
		return fmt.Errorf("derak: get zone ID: %w", err)
	}

	r := internal.Record{
		Type:    "TXT",
		Host:    recordName,
		Content: info.Value,
		TTL:     d.config.TTL,
	}

	record, err := d.client.CreateRecord(ctx, zoneID, r)
	if err != nil {
		return fmt.Errorf("derak: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = record.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneID, err := d.getZoneID(ctx, info)
	if err != nil {
		return fmt.Errorf("derak: get zone ID: %w", err)
	}

	// gets the record's unique ID
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("derak: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err = d.client.DeleteRecord(ctx, zoneID, recordID)
	if err != nil {
		return fmt.Errorf("derak: delete record: %w", err)
	}

	// deletes record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) getZoneID(ctx context.Context, info dns01.ChallengeInfo) (string, error) {
	zoneID := d.config.WebsiteID
	if zoneID != "" {
		return zoneID, nil
	}

	zones, err := d.client.GetZones(ctx)
	if err != nil {
		return "", fmt.Errorf("get zones: %w", err)
	}

	for _, zone := range zones {
		if strings.HasSuffix(info.EffectiveFQDN, dns.Fqdn(zone.HumanReadable)) {
			return zone.ID, nil
		}
	}

	return "", fmt.Errorf("zone/website not found %s", info.EffectiveFQDN)
}
