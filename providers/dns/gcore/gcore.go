// Package gcore implements a DNS provider for solving the DNS-01 challenge using G-Core.
package gcore

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/gcore"
)

// Environment variables names.
const (
	envNamespace = "GCORE_"

	EnvPermanentAPIToken = envNamespace + "PERMANENT_API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config for DNSProvider.
type Config = gcore.Config

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, gcore.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, gcore.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}
}

// DNSProvider an implementation of challenge.Provider contract.
type DNSProvider struct {
	prv challenge.ProviderTimeout
}

// NewDNSProvider returns an instance of DNSProvider configured for G-Core DNS API.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvPermanentAPIToken)
	if err != nil {
		return nil, fmt.Errorf("gcore: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvPermanentAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for G-Core DNS API.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gcore: the configuration of the DNS provider is nil")
	}

	provider, err := gcore.NewDNSProviderConfig(config, "")
	if err != nil {
		return nil, fmt.Errorf("gcore: %w", err)
	}

	return &DNSProvider{prv: provider}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	err := d.prv.Present(domain, token, keyAuth)
	if err != nil {
		return fmt.Errorf("gcore: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	err := d.prv.CleanUp(domain, token, keyAuth)
	if err != nil {
		return fmt.Errorf("gcore: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.prv.Timeout()
}
