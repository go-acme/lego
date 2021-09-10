// Package softlayer implements a DNS provider for solving the DNS-01 challenge using IBM Cloud Domain Name Registration.
package softlayer

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/softlayer/softlayer-go/session"
)

// Environment variables names.
const (
	envNamespace = "SOFTLAYER_"

	EnvUsername = envNamespace + "USERNAME"
	EnvAPIKey   = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"

	EnvDebug = envNamespace + "DEBUG"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username           string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
	Debug              bool
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
		Debug: env.GetOrDefaultBool(EnvDebug, false),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *session.Session
}

// NewDNSProvider returns a DNSProvider instance configured for softlayer.
// Credentials must be passed in the environment variables:
// SOFTLAYER_USERNAME & SOFTLAYER_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("softlayer: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for softlayer.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("softlayer: the configuration of the DNS provider is nil")
	}

	if config.Username == "" {
		return nil, errors.New("softlayer: Username is missing")
	}

	if config.APIKey == "" {
		return nil, errors.New("softlayer: Api key is missing")
	}

	apiClient := session.New(config.Username, config.APIKey)
	if config.HTTPClient == nil {
		apiClient.HTTPClient = http.DefaultClient
	} else {
		apiClient.HTTPClient = config.HTTPClient
	}
	apiClient.Debug = config.Debug

	return &DNSProvider{
		client: apiClient,
		config: config,
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.addTXTRecord(fqdn, domain, value, d.config.TTL)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	return d.cleanupTXTRecord(fqdn, domain)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
