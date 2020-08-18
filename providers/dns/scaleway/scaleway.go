// Package scaleway implements a DNS provider for solving the DNS-01 challenge using Scaleway Domains API.
// Scaleway Domain API reference: https://developers.scaleway.com/en/products/domain/api/
// Token: https://www.scaleway.com/en/docs/generate-an-api-token/
package scaleway

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/providers/dns/scaleway/internal"
)

const (
	defaultBaseURL            = "https://api.scaleway.com"
	defaultVersion            = "v2alpha2"
	minTTL                    = 60
	defaultPollingInterval    = 10 * time.Second
	defaultPropagationTimeout = 120 * time.Second
)

// Environment variables names.
const (
	envNamespace = "SCALEWAY_"

	EnvBaseURL    = envNamespace + "BASE_URL"
	EnvAPIToken   = envNamespace + "API_TOKEN"
	EnvAPIVersion = envNamespace + "API_VERSION"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	Version            string
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		BaseURL:            env.GetOrDefaultString(conf, EnvBaseURL, defaultBaseURL),
		Version:            env.GetOrDefaultString(conf, EnvAPIVersion, defaultVersion),
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, defaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Scaleway Domains API.
// API token must be passed in the environment variable SCALEWAY_API_TOKEN.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("scaleway: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.Token = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for scaleway.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("scaleway: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("scaleway: credentials missing")
	}

	if config.TTL < minTTL {
		config.TTL = minTTL
	}

	client := internal.NewClient(internal.ClientOpts{
		BaseURL: fmt.Sprintf("%s/domain/%s", config.BaseURL, config.Version),
		Token:   config.Token,
	}, config.HTTPClient)

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

	txtRecord := internal.Record{
		Type: "TXT",
		TTL:  uint32(d.config.TTL),
		Name: fqdn,
		Data: fmt.Sprintf(`"%s"`, value),
	}

	err := d.client.AddRecord(domain, txtRecord)
	if err != nil {
		return fmt.Errorf("scaleway: %w", err)
	}
	return nil
}

// CleanUp removes a TXT record used for DNS-01 challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	txtRecord := internal.Record{
		Type: "TXT",
		TTL:  uint32(d.config.TTL),
		Name: fqdn,
		Data: fmt.Sprintf(`"%s"`, value),
	}

	err := d.client.DeleteRecord(domain, txtRecord)
	if err != nil {
		return fmt.Errorf("scaleway: %w", err)
	}
	return nil
}
