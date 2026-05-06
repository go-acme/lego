// Package excedo implements a DNS provider for solving the DNS-01 challenge using Excedo.
package excedo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/excedo/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "EXCEDO_"

	EnvAPIURL = envNamespace + "API_URL"
	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIURL string
	APIKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
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

	recordsMu sync.Mutex
	records   map[string]int64
}

// NewDNSProvider returns a DNSProvider instance configured for Excedo.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIURL, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("excedo: %w", err)
	}

	config := NewDefaultConfig()
	config.APIURL = values[EnvAPIURL]
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Excedo.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("excedo: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIURL, config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("excedo: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:  config,
		client:  client,
		records: make(map[string]int64),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("excedo: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("excedo: %w", err)
	}

	record := internal.Record{
		DomainName: dns01.UnFqdn(authZone),
		Name:       subDomain,
		Type:       "TXT",
		Content:    info.Value,
		TTL:        strconv.Itoa(d.config.TTL),
	}

	recordID, err := d.client.AddRecord(ctx, record)
	if err != nil {
		return fmt.Errorf("excedo: add record: %w", err)
	}

	d.recordsMu.Lock()
	d.records[token] = recordID
	d.recordsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("excedo: could not find zone for domain %q: %w", domain, err)
	}

	d.recordsMu.Lock()
	recordID, ok := d.records[token]
	d.recordsMu.Unlock()

	if !ok {
		return fmt.Errorf("excedo: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	err = d.client.DeleteRecord(ctx, dns01.UnFqdn(authZone), strconv.FormatInt(recordID, 10))
	if err != nil {
		return fmt.Errorf("excedo: delete record: %w", err)
	}

	d.recordsMu.Lock()
	delete(d.records, token)
	d.recordsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
