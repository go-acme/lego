// Package dnscale implements a DNS provider for solving the DNS-01 challenge using the DNScale API.
package dnscale

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/dnscale/internal"
)

// Environment variables names.
const (
	envNamespace = "DNSCALE_"

	EnvAPIToken = envNamespace + "API_TOKEN"
	EnvAPIURL   = envNamespace + "API_URL"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const defaultBaseURL = "https://api.dnscale.eu"

// Ensure DNSProvider implements the required interfaces.
var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIToken           string
	BaseURL            string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            env.GetOrDefaultString(EnvAPIURL, defaultBaseURL),
		TTL:                env.GetOrDefaultInt(EnvTTL, 120),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 5*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for DNScale.
// Credentials must be passed in the environment variable: DNSCALE_API_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("dnscale: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig returns a DNSProvider instance configured for DNScale.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnscale: the configuration of the DNS provider is nil")
	}

	if config.APIToken == "" {
		return nil, errors.New("dnscale: some credentials information are missing: DNSCALE_API_TOKEN")
	}

	client := internal.NewClient(config.BaseURL, config.APIToken)
	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneID, _, err := d.client.FindZoneByFQDN(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dnscale: could not find zone for domain %q: %w", domain, err)
	}

	recordName := dns01.UnFqdn(info.EffectiveFQDN)

	err = d.client.CreateTXTRecord(ctx, zoneID, recordName, info.Value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("dnscale: create TXT record for %s: %w", recordName, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneID, _, err := d.client.FindZoneByFQDN(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dnscale: could not find zone for domain %q: %w", domain, err)
	}

	recordName := dns01.UnFqdn(info.EffectiveFQDN)

	err = d.client.DeleteTXTRecord(ctx, zoneID, recordName, info.Value)
	if err != nil {
		return fmt.Errorf("dnscale: delete TXT record for %s: %w", recordName, err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
