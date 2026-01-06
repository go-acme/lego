// Package com35 implements a DNS provider for solving the DNS-01 challenge using 35.com/三五互联.
package com35

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/westcn"
)

// Environment variables names.
const (
	envNamespace = "COM35_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const defaultBaseURL = "https://api.35.cn/api/v2"

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config = westcn.Config

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	prv challenge.ProviderTimeout
}

// NewDNSProvider returns a DNSProvider instance configured for 35.com/三五互联.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("35com: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for 35.com/三五互联.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("35com: the configuration of the DNS provider is nil")
	}

	provider, err := westcn.NewDNSProviderConfig(config, defaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("35com: %w", err)
	}

	return &DNSProvider{prv: provider}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	err := d.prv.Present(domain, token, keyAuth)
	if err != nil {
		return fmt.Errorf("35com: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	err := d.prv.CleanUp(domain, token, keyAuth)
	if err != nil {
		return fmt.Errorf("35com: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.prv.Timeout()
}
