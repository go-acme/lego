// Package zoneee implements a DNS provider for solving the DNS-01 challenge through zone.ee.
package zoneee

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/zoneee/internal"
)

// Environment variables names.
const (
	envNamespace = "ZONEEE_"

	EnvEndpoint = envNamespace + "ENDPOINT"
	EnvAPIUser  = envNamespace + "API_USER"
	EnvAPIKey   = envNamespace + "API_KEY"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Endpoint           *url.URL
	Username           string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	endpoint, _ := url.Parse(internal.DefaultEndpoint)

	return &Config{
		Endpoint: endpoint,
		// zone.ee can take up to 5min to propagate according to the support
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 5*time.Minute),
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
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIUser, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("zoneee: %w", err)
	}

	rawEndpoint := env.GetOrDefaultString(EnvEndpoint, internal.DefaultEndpoint)
	endpoint, err := url.Parse(rawEndpoint)
	if err != nil {
		return nil, fmt.Errorf("zoneee: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvAPIUser]
	config.APIKey = values[EnvAPIKey]
	config.Endpoint = endpoint

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Zone.ee.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("zoneee: the configuration of the DNS provider is nil")
	}

	if config.Username == "" {
		return nil, errors.New("zoneee: credentials missing: username")
	}

	if config.APIKey == "" {
		return nil, errors.New("zoneee: credentials missing: API key")
	}

	if config.Endpoint == nil {
		return nil, errors.New("zoneee: the endpoint is missing")
	}

	client := internal.NewClient(config.Username, config.APIKey)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	if config.Endpoint != nil {
		client.BaseURL = config.Endpoint
	}

	return &DNSProvider{config: config, client: client}, nil
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
		return fmt.Errorf("zoneee: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	record := internal.TXTRecord{
		Name:        dns01.UnFqdn(info.EffectiveFQDN),
		Destination: info.Value,
	}

	_, err = d.client.AddTxtRecord(context.Background(), authZone, record)
	if err != nil {
		return fmt.Errorf("zoneee: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record previously created.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("zoneee: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	ctx := context.Background()

	records, err := d.client.GetTxtRecords(ctx, authZone)
	if err != nil {
		return fmt.Errorf("zoneee: %w", err)
	}

	var id string
	for _, record := range records {
		if record.Destination == info.Value {
			id = record.ID
		}
	}

	if id == "" {
		return fmt.Errorf("zoneee: txt record does not exist for %s", info.Value)
	}

	if err = d.client.RemoveTxtRecord(ctx, authZone, id); err != nil {
		return fmt.Errorf("zoneee: %w", err)
	}

	return nil
}
