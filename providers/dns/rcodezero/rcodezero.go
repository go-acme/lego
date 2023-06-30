// Package rcodezero implements a DNS provider for solving the DNS-01 challenge using RcodeZero Anycast network.
package rcodezero

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/rcodezero/internal"
)

// Environment variables names.
const (
	envNamespace = "RCODEZERO_"

	EnvAPIToken = envNamespace + "API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIToken           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 240*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for RcodeZero.
// Credentials must be passed in the environment variable:
// RCODEZERO_API_URL and RCODEZERO_API_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("rcodezero: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for RcodeZero.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("rcodezero: the configuration of the DNS provider is nil")
	}

	if config.APIToken == "" {
		return nil, errors.New("rcodezero: API token missing")
	}

	client := internal.NewClient(config.APIToken)

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

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("rcodezero: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	rrSet := []internal.UpdateRRSet{{
		Name:       info.EffectiveFQDN,
		ChangeType: "update",
		Type:       "TXT",
		TTL:        d.config.TTL,
		Records:    []internal.Record{{Content: `"` + info.Value + `"`}},
	}}

	_, err = d.client.UpdateRecords(ctx, authZone, rrSet)
	if err != nil {
		return fmt.Errorf("rcodezero: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("rcodezero: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	rrSet := []internal.UpdateRRSet{{
		Name:       info.EffectiveFQDN,
		Type:       "TXT",
		ChangeType: "delete",
	}}

	_, err = d.client.UpdateRecords(ctx, authZone, rrSet)
	if err != nil {
		return fmt.Errorf("rcodezero: %w", err)
	}

	return nil
}
