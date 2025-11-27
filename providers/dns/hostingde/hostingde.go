// Package hostingde implements a DNS provider for solving the DNS-01 challenge using hosting.de.
package hostingde

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/hostingde"
)

// Environment variables names.
const (
	envNamespace = "HOSTINGDE_"

	EnvAPIKey   = envNamespace + "API_KEY"
	EnvZoneName = envNamespace + "ZONE_NAME"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config = hostingde.Config

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		ZoneName:           env.GetOrFile(EnvZoneName),
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
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

// NewDNSProvider returns a DNSProvider instance configured for hosting.de.
// Credentials must be passed in the environment variables:
// HOSTINGDE_ZONE_NAME and HOSTINGDE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("hostingde: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for hosting.de.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("hostingde: the configuration of the DNS provider is nil")
	}

	provider, err := hostingde.NewDNSProviderConfig(config, "")
	if err != nil {
		return nil, fmt.Errorf("hostingde: %w", err)
	}

	return &DNSProvider{prv: provider}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	err := d.prv.Present(domain, token, keyAuth)
	if err != nil {
		return fmt.Errorf("hostingde: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	err := d.prv.CleanUp(domain, token, keyAuth)
	if err != nil {
		return fmt.Errorf("hostingde: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.prv.Timeout()
}
