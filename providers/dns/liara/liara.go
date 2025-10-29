// Package liara implements a DNS provider for solving the DNS-01 challenge using Liara DNS.
package liara

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/liara/internal"
	"github.com/hashicorp/go-retryablehttp"
)

// Environment variables names.
const (
	envNamespace = "LIARA_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const (
	minTTL = 120
	maxTTL = 432000
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
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

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Liara DNS.
// Liara_API_KEY must be passed in the environment variables.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("liara: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Liara DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("liara: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("liara: APIKey is missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("liara: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	if config.TTL > maxTTL {
		return nil, fmt.Errorf("liara: invalid TTL, TTL (%d) must be lower than %d", config.TTL, maxTTL)
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	if config.HTTPClient != nil {
		retryClient.HTTPClient = config.HTTPClient
	}
	retryClient.Logger = log.Logger

	client := internal.NewClient(
		clientdebug.Wrap(
			internal.OAuthStaticAccessToken(retryClient.StandardClient(), config.APIKey),
		),
	)

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

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("liara: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("liara: %w", err)
	}

	record := internal.Record{
		Type:     "TXT",
		Name:     subDomain,
		Contents: []internal.Content{{Text: info.Value}},
		TTL:      d.config.TTL,
	}
	newRecord, err := d.client.CreateRecord(context.Background(), dns01.UnFqdn(authZone), record)
	if err != nil {
		return fmt.Errorf("liara: failed to create TXT record, fqdn=%s: %w", info.EffectiveFQDN, err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("liara: could not find zone for domain %q: %w", domain, err)
	}

	// gets the record's unique ID
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("liara: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err = d.client.DeleteRecord(context.Background(), dns01.UnFqdn(authZone), recordID)
	if err != nil {
		return fmt.Errorf("liara: failed to delete TXT record, id=%s: %w", recordID, err)
	}

	// deletes record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}
