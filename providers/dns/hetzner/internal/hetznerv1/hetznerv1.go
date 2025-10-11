// Package hetznerv1 implements a DNS provider for solving the DNS-01 challenge using Hetzner.
package hetznerv1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/platform/wait"
	"github.com/go-acme/lego/v4/providers/dns/hetzner/internal/hetznerv1/internal"
)

// Environment variables names.
const (
	envNamespace = "HETZNER_"

	EnvAPIToken = envNamespace + "API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIToken string

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

// NewDNSProvider returns a DNSProvider instance configured for Hetzner.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("hetzner: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Hetzner.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("hetzner: the configuration of the DNS provider is nil")
	}

	if config.APIToken == "" {
		return nil, errors.New("hetzner: credentials missing")
	}

	client, err := internal.NewClient(internal.OAuthStaticAccessToken(config.HTTPClient, config.APIToken))
	if err != nil {
		return nil, fmt.Errorf("hetzner: %w", err)
	}

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hetzner: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("hetzner: %w", err)
	}

	records := []internal.Record{{Value: strconv.Quote(info.Value)}}

	action, err := d.client.AddRRSetRecords(ctx, dns01.UnFqdn(authZone), "TXT", subDomain, d.config.TTL, records)
	if err != nil {
		return fmt.Errorf("hetzner: add RRSet records: %w", err)
	}

	return d.waitAction(ctx, "add RRSet records", action.ID)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hetzner: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("hetzner: %w", err)
	}

	records := []internal.Record{{Value: strconv.Quote(info.Value)}}

	action, err := d.client.RemoveRRSetRecords(ctx, dns01.UnFqdn(authZone), "TXT", subDomain, records)
	if err != nil {
		return fmt.Errorf("hetzner: remove RRSet records: %w", err)
	}

	return d.waitAction(ctx, "remove RRSet records", action.ID)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) waitAction(ctx context.Context, msg string, actionID int) error {
	return wait.For(msg, d.config.PropagationTimeout, d.config.PollingInterval, func() (bool, error) {
		result, err := d.client.GetAction(ctx, actionID)
		if err != nil {
			return false, fmt.Errorf("hetzner: get action: %w", err)
		}

		if result.Progress < 100 {
			return false, nil
		}

		if result.ErrorInfo != nil {
			return false, fmt.Errorf("hetzner: %s", result.ErrorInfo.Message)
		}

		return true, nil
	})
}
