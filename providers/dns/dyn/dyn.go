// Package dyn implements a DNS provider for solving the DNS-01 challenge using Dyn Managed DNS.
package dyn

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/dyn/internal"
)

// Environment variables names.
const (
	envNamespace = "DYN_"

	EnvCustomerName = envNamespace + "CUSTOMER_NAME"
	EnvUserName     = envNamespace + "USER_NAME"
	EnvPassword     = envNamespace + "PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	CustomerName       string
	UserName           string
	Password           string
	HTTPClient         *http.Client
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
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
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Dyn DNS.
// Credentials must be passed in the environment variables:
// DYN_CUSTOMER_NAME, DYN_USER_NAME and DYN_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvCustomerName, EnvUserName, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("dyn: %w", err)
	}

	config := NewDefaultConfig()
	config.CustomerName = values[EnvCustomerName]
	config.UserName = values[EnvUserName]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Dyn DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dyn: the configuration of the DNS provider is nil")
	}

	if config.CustomerName == "" || config.UserName == "" || config.Password == "" {
		return nil, errors.New("dyn: credentials missing")
	}

	client := internal.NewClient(config.CustomerName, config.UserName, config.Password)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dyn: could not find zone for domain %q: %w", domain, err)
	}

	ctx, err := d.client.CreateAuthenticatedContext(context.Background())
	if err != nil {
		return fmt.Errorf("dyn: %w", err)
	}

	err = d.client.AddTXTRecord(ctx, authZone, info.EffectiveFQDN, info.Value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("dyn: %w", err)
	}

	err = d.client.Publish(ctx, authZone, "Added TXT record for ACME dns-01 challenge using lego client")
	if err != nil {
		return fmt.Errorf("dyn: %w", err)
	}

	return d.client.Logout(ctx)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dyn: could not find zone for domain %q: %w", domain, err)
	}

	ctx, err := d.client.CreateAuthenticatedContext(context.Background())
	if err != nil {
		return fmt.Errorf("dyn: %w", err)
	}

	err = d.client.RemoveTXTRecord(ctx, authZone, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dyn: %w", err)
	}

	err = d.client.Publish(ctx, authZone, "Removed TXT record for ACME dns-01 challenge using lego client")
	if err != nil {
		return fmt.Errorf("dyn: %w", err)
	}

	return d.client.Logout(ctx)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
