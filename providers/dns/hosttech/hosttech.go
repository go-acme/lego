// Package hosttech implements a DNS provider for solving the DNS-01 challenge using hosttech.
package hosttech

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/hosttech/internal"
)

// Environment variables names.
const (
	envNamespace = "HOSTTECH_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

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
		TTL:                env.GetOrDefaultInt(EnvTTL, 3600),
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

	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for hosttech.
// Credentials must be passed in the environment variable: HOSTTECH_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("hosttech: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for hosttech.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("hosttech: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("hosttech: missing credentials")
	}

	client := internal.NewClient(internal.OAuthStaticAccessToken(config.HTTPClient, config.APIKey))

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: map[string]int{},
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
		return fmt.Errorf("hosttech: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	zone, err := d.client.GetZone(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("hosttech: could not find zone for domain %q (%s): %w", domain, authZone, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("hosttech: %w", err)
	}

	record := internal.Record{
		Type: "TXT",
		Name: subDomain,
		Text: info.Value,
		TTL:  d.config.TTL,
	}

	newRecord, err := d.client.AddRecord(ctx, strconv.Itoa(zone.ID), record)
	if err != nil {
		return fmt.Errorf("hosttech: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hosttech: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	zone, err := d.client.GetZone(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("hosttech: could not find zone for domain %q (%s): %w", domain, authZone, err)
	}

	// gets the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("hosttech: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err = d.client.DeleteRecord(ctx, strconv.Itoa(zone.ID), strconv.Itoa(recordID))
	if err != nil {
		return fmt.Errorf("hosttech: %w", err)
	}

	return nil
}
