// Package websupport implements a DNS provider for solving the DNS-01 challenge using Websupport.
package websupport

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/active24"
)

const baseAPIDomain = "websupport.sk"

// Environment variables names.
const (
	envNamespace = "WEBSUPPORT_"

	EnvAPIKey = envNamespace + "API_KEY"
	EnvSecret = envNamespace + "SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config = active24.Config

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
	prv challenge.ProviderTimeout
}

// NewDNSProvider returns a DNSProvider instance configured for Websupport.
// Credentials must be passed in the environment variables: WEBSUPPORT_API_KEY, WEBSUPPORT_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvSecret)
	if err != nil {
		return nil, fmt.Errorf("websupport: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.Secret = values[EnvSecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Websupport.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("websupport: the configuration of the DNS provider is nil")
	}

	provider, err := active24.NewDNSProviderConfig(config, baseAPIDomain)
	if err != nil {
		return nil, fmt.Errorf("websupport: %w", err)
	}

	return &DNSProvider{prv: provider}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	err := d.prv.Present(domain, token, keyAuth)
	if err != nil {
		return fmt.Errorf("websupport: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	err := d.prv.CleanUp(domain, token, keyAuth)
	if err != nil {
		return fmt.Errorf("websupport: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.prv.Timeout()
}
