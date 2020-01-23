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

const (
	envNamespace             = "SCALEWAY_"
	baseURLEnvVar            = envNamespace + "BASE_URL"
	apiTokenEnvVar           = envNamespace + "API_TOKEN"
	apiVersionEnvVar         = envNamespace + "API_VERSION"
	ttlEnvVar                = envNamespace + "TTL"
	propagationTimeoutEnvVar = envNamespace + "PROPAGATION_TIMEOUT"
	pollingIntervalEnvVar    = envNamespace + "POLLING_INTERVAL"
	httpTimeoutEnvVar        = envNamespace + "HTTP_TIMEOUT"
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
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            env.GetOrDefaultString(baseURLEnvVar, defaultBaseURL),
		Version:            env.GetOrDefaultString(apiVersionEnvVar, defaultVersion),
		TTL:                env.GetOrDefaultInt(ttlEnvVar, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(propagationTimeoutEnvVar, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(pollingIntervalEnvVar, defaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(httpTimeoutEnvVar, 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Scaleway Domains API.
// API token must be passed in the environment variable SCALEWAY_API_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(apiTokenEnvVar)
	if err != nil {
		return nil, fmt.Errorf("scaleway: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[apiTokenEnvVar]

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
