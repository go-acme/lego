// Package resellerclub implements a DNS provider for solving the DNS-01 challenge using ResellerClub.
package resellerclub

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/resellerclub/internal"
)

// Environment variables names.
const (
	envNamespace = "RESELLERCLUB_"

	EnvAPIKey = envNamespace + "API_KEY"
	EnvUserID = envNamespace + "USER_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey string
	UserID string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

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
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for ResellerClub.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvUserID)
	if err != nil {
		return nil, fmt.Errorf("resellerclub: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.UserID = values[EnvUserID]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ResellerClub.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("resellerclub: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIKey, config.UserID)
	if err != nil {
		return nil, fmt.Errorf("resellerclub: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("resellerclub: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("resellerclub: %w", err)
	}

	record := internal.Record{
		Type:  "TXT",
		Value: info.Value,
		Host:  subDomain,
		TTL:   d.config.TTL,
	}

	_, err = d.client.CreateDNSRecord(ctx, dns01.UnFqdn(authZone), record)
	if err != nil {
		return fmt.Errorf("resellerclub: create DNS record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("resellerclub: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("resellerclub: %w", err)
	}

	record := internal.Record{
		Type:  "TXT",
		Value: info.Value,
		Host:  subDomain,
	}

	_, err = d.client.DeleteDNSRecord(ctx, dns01.UnFqdn(authZone), record)
	if err != nil {
		return fmt.Errorf("resellerclub: delete DNS record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
