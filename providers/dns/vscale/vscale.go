// Package vscale implements a DNS provider for solving the DNS-01 challenge using Vscale Domains API.
// Vscale Domain API reference: https://developers.vscale.io/documentation/api/v1/#api-Domains
// Token: https://vscale.io/panel/settings/tokens/
package vscale

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/providers/dns/internal/selectel"
)

const minTTL = 60

// Environment variables names.
const (
	envNamespace = "VSCALE_"

	EnvBaseURL  = envNamespace + "BASE_URL"
	EnvAPIToken = envNamespace + "API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		BaseURL:            env.GetOrDefaultString(conf, EnvBaseURL, selectel.DefaultVScaleBaseURL),
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *selectel.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Vscale Domains API.
// API token must be passed in the environment variable VSCALE_API_TOKEN.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("vscale: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.Token = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Vscale.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("vscale: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("vscale: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("vscale: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := selectel.NewClient(config.Token)
	client.BaseURL = config.BaseURL
	client.HTTPClient = config.HTTPClient

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the Timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill DNS-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	domainObj, err := d.client.GetDomainByName(domain)
	if err != nil {
		return fmt.Errorf("vscale: %w", err)
	}

	txtRecord := selectel.Record{
		Type:    "TXT",
		TTL:     d.config.TTL,
		Name:    fqdn,
		Content: value,
	}
	_, err = d.client.AddRecord(domainObj.ID, txtRecord)
	if err != nil {
		return fmt.Errorf("vscale: %w", err)
	}

	return nil
}

// CleanUp removes a TXT record used for DNS-01 challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	recordName := dns01.UnFqdn(fqdn)

	domainObj, err := d.client.GetDomainByName(domain)
	if err != nil {
		return fmt.Errorf("vscale: %w", err)
	}

	records, err := d.client.ListRecords(domainObj.ID)
	if err != nil {
		return fmt.Errorf("vscale: %w", err)
	}

	// Delete records with specific FQDN
	var lastErr error
	for _, record := range records {
		if record.Name == recordName {
			err = d.client.DeleteRecord(domainObj.ID, record.ID)
			if err != nil {
				lastErr = fmt.Errorf("vscale: %w", err)
			}
		}
	}

	return lastErr
}
