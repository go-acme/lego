// Package leaseweb implements a DNS provider for solving the DNS-01 challenge using Leaseweb.
package leaseweb

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/internal/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/leaseweb/internal"
)

// Environment variables names.
const (
	envNamespace = "LEASEWEB_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey string

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

// NewDNSProvider returns a DNSProvider instance configured for Leaseweb.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("leaseweb: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Leaseweb.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("leaseweb: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("leaseweb: %w", err)
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
		return fmt.Errorf("leaseweb: could not find zone for domain %q: %w", domain, err)
	}

	existingRRSet, err := d.client.GetRRSet(ctx, dns01.UnFqdn(authZone), info.EffectiveFQDN, "TXT")
	if err != nil {
		notfoundErr := &internal.NotFoundError{}
		if !errors.As(err, &notfoundErr) {
			return fmt.Errorf("leaseweb: get RRSet: %w", err)
		}

		// Create the RRSet.

		rrset := internal.RRSet{
			Content: []string{info.Value},
			Name:    info.EffectiveFQDN,
			TTL:     internal.TTLRounder(d.config.TTL),
			Type:    "TXT",
		}

		_, err = d.client.CreateRRSet(ctx, dns01.UnFqdn(authZone), rrset)
		if err != nil {
			return fmt.Errorf("leaseweb: create RRSet: %w", err)
		}

		return nil
	}

	// Update the RRSet.

	existingRRSet.Content = append(existingRRSet.Content, info.Value)

	_, err = d.client.UpdateRRSet(ctx, dns01.UnFqdn(authZone), *existingRRSet)
	if err != nil {
		return fmt.Errorf("leaseweb: update RRSet: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("leaseweb: could not find zone for domain %q: %w", domain, err)
	}

	existingRRSet, err := d.client.GetRRSet(ctx, dns01.UnFqdn(authZone), info.EffectiveFQDN, "TXT")
	if err != nil {
		return fmt.Errorf("leaseweb: get RRSet: %w", err)
	}

	var content []string

	for _, s := range existingRRSet.Content {
		if s != info.Value {
			content = append(content, s)
		}
	}

	if len(content) == 0 {
		err = d.client.DeleteRRSet(ctx, dns01.UnFqdn(authZone), info.EffectiveFQDN, "TXT")
		if err != nil {
			return fmt.Errorf("leaseweb: delete RRSet: %w", err)
		}

		return nil
	}

	existingRRSet.Content = content

	_, err = d.client.UpdateRRSet(ctx, dns01.UnFqdn(authZone), *existingRRSet)
	if err != nil {
		return fmt.Errorf("leaseweb: update RRSet: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
