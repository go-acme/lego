// Package technitium implements a DNS provider for solving the DNS-01 challenge using Technitium.
package technitium

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/technitium/internal"
)

// Environment variables names.
const (
	envNamespace = "TECHNITIUM_"

	EnvServerBaseURL = envNamespace + "SERVER_BASE_URL"
	EnvAPIToken      = envNamespace + "API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL  string
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
}

// NewDNSProvider returns a DNSProvider instance configured for Technitium.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvServerBaseURL, EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("technitium: %w", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values[EnvServerBaseURL]
	config.APIToken = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Technitium.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("technitium: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.BaseURL, config.APIToken)
	if err != nil {
		return nil, fmt.Errorf("technitium: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	record := internal.Record{
		Domain: info.EffectiveFQDN,
		Type:   "TXT",
		Text:   info.Value,
	}

	_, err := d.client.AddRecord(context.Background(), record)
	if err != nil {
		return fmt.Errorf("technitium: add record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	record := internal.Record{
		Domain: info.EffectiveFQDN,
		Type:   "TXT",
		Text:   info.Value,
	}

	err := d.client.DeleteRecord(context.Background(), record)
	if err != nil {
		return fmt.Errorf("technitium: delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
