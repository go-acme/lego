// Package rimuhosting implements a DNS provider for solving the DNS-01 challenge using RimuHosting DNS.
package rimuhosting

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/rimuhosting"
)

// Environment variables names.
const (
	envNamespace = "RIMUHOSTING_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey string

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
	client *rimuhosting.Client
}

// NewDNSProvider returns a DNSProvider instance configured for RimuHosting.
// Credentials must be passed in the environment variables.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("rimuhosting: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for RimuHosting.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("rimuhosting: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("rimuhosting: incomplete credentials, missing API key")
	}

	client := rimuhosting.NewClient(config.APIKey)
	client.BaseURL = rimuhosting.DefaultRimuHostingBaseURL

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	records, err := d.client.FindTXTRecords(dns01.UnFqdn(fqdn))
	if err != nil {
		return fmt.Errorf("rimuhosting: failed to find record(s) for %s: %w", domain, err)
	}

	actions := []rimuhosting.ActionParameter{
		rimuhosting.AddRecord(dns01.UnFqdn(fqdn), value, d.config.TTL),
	}

	for _, record := range records {
		actions = append(actions, rimuhosting.AddRecord(record.Name, record.Content, d.config.TTL))
	}

	_, err = d.client.DoActions(actions...)
	if err != nil {
		return fmt.Errorf("rimuhosting: failed to add record(s) for %s: %w", domain, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	action := rimuhosting.DeleteRecord(dns01.UnFqdn(fqdn), value)

	_, err := d.client.DoActions(action)
	if err != nil {
		return fmt.Errorf("rimuhosting: failed to delete record for %s: %w", domain, err)
	}

	return nil
}
