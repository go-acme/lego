// Package domeneshop implements a DNS provider for solving the DNS-01 challenge using domeneshop DNS.
package domeneshop

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/domeneshop/internal"
)

// Environment variables names.
const (
	envNamespace = "DOMENESHOP_"

	EnvAPIToken  = envNamespace + "API_TOKEN"
	EnvAPISecret = envNamespace + "API_SECRET"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIToken           string
	APISecret          string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 5*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 20*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for domeneshop.
// Credentials must be passed in the environment variables:
// DOMENESHOP_API_TOKEN, DOMENESHOP_API_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("domeneshop: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvAPIToken]
	config.APISecret = values[EnvAPISecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Domeneshop.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("domeneshop: the configuration of the DNS provider is nil")
	}

	if config.APIToken == "" || config.APISecret == "" {
		return nil, errors.New("domeneshop: credentials missing")
	}

	client := internal.NewClient(config.APIToken, config.APISecret)

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
func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, host, err := d.splitDomain(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("domeneshop: %w", err)
	}

	ctx := context.Background()

	domainInstance, err := d.client.GetDomainByName(ctx, zone)
	if err != nil {
		return fmt.Errorf("domeneshop: %w", err)
	}

	err = d.client.CreateTXTRecord(ctx, domainInstance, host, info.Value)
	if err != nil {
		return fmt.Errorf("domeneshop: failed to create record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, host, err := d.splitDomain(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("domeneshop: %w", err)
	}

	ctx := context.Background()

	domainInstance, err := d.client.GetDomainByName(ctx, zone)
	if err != nil {
		return fmt.Errorf("domeneshop: %w", err)
	}

	if err := d.client.DeleteTXTRecord(ctx, domainInstance, host, info.Value); err != nil {
		return fmt.Errorf("domeneshop: failed to create record: %w", err)
	}

	return nil
}

// splitDomain splits the hostname from the authoritative zone, and returns both parts (non-fqdn).
func (d *DNSProvider) splitDomain(fqdn string) (string, string, error) {
	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", "", fmt.Errorf("could not find zone: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, zone)
	if err != nil {
		return "", "", err
	}

	return dns01.UnFqdn(zone), subDomain, nil
}
