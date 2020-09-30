// Package dreamhost implements a DNS provider for solving the DNS-01 challenge using DreamHost.
// See https://help.dreamhost.com/hc/en-us/articles/217560167-API_overview
// and https://help.dreamhost.com/hc/en-us/articles/217555707-DNS-API-commands for the API spec.
package dreamhost

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "DREAMHOST_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 60*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 1*time.Minute),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a new DNS provider using
// environment variable DREAMHOST_API_KEY for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("dreamhost: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DreamHost.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dreamhost: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("dreamhost: credentials missing")
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.CreateRecord(domain, token, fqdn, value)
}

// CreateRecord creates a TXT record to fulfill the DNS-01 challenge.
func (d *DNSProvider) CreateRecord(domain, token, fqdn, value string) error {
	record := dns01.UnFqdn(fqdn)

	u, err := d.buildQuery(cmdAddRecord, record, value)
	if err != nil {
		return fmt.Errorf("dreamhost: %w", err)
	}

	err = d.updateTxtRecord(u)
	if err != nil {
		return fmt.Errorf("dreamhost: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.DeleteRecord(domain, token, fqdn, value)
}

// DeleteRecord removes a creates a TXT record from the provider.
func (d *DNSProvider) DeleteRecord(domain, token, fqdn, value string) error {
	record := dns01.UnFqdn(fqdn)

	u, err := d.buildQuery(cmdRemoveRecord, record, value)
	if err != nil {
		return fmt.Errorf("dreamhost: %w", err)
	}

	err = d.updateTxtRecord(u)
	if err != nil {
		return fmt.Errorf("dreamhost: %w", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
