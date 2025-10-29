// Package metaregistrar implements a DNS provider for solving the DNS-01 challenge using Metaregistrar.
package metaregistrar

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/metaregistrar/internal"
)

// Environment variables names.
const (
	envNamespace = "METAREGISTRAR_"

	EnvToken = envNamespace + "API_TOKEN"

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

// NewDNSProvider returns a DNSProvider instance configured for Metaregistrar.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken)
	if err != nil {
		return nil, fmt.Errorf("metaregistrar: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Metaregistrar.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("metaregistrar: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIToken)
	if err != nil {
		return nil, fmt.Errorf("metaregistrar: %w", err)
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
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("metaregistrar: could not find zone for domain %q: %w", domain, err)
	}

	updateRequest := internal.DNSZoneUpdateRequest{
		Add: []internal.Record{{
			Name:    dns01.UnFqdn(info.EffectiveFQDN),
			Type:    "TXT",
			TTL:     d.config.TTL,
			Content: info.Value,
		}},
	}

	_, err = d.client.UpdateDNSZone(context.Background(), dns01.UnFqdn(authZone), updateRequest)
	if err != nil {
		return fmt.Errorf("metaregistrar: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("metaregistrar: could not find zone for domain %q: %w", domain, err)
	}

	updateRequest := internal.DNSZoneUpdateRequest{
		Remove: []internal.Record{{
			Name:    dns01.UnFqdn(info.EffectiveFQDN),
			Type:    "TXT",
			TTL:     d.config.TTL,
			Content: strconv.Quote(info.Value),
		}},
	}

	_, err = d.client.UpdateDNSZone(context.Background(), dns01.UnFqdn(authZone), updateRequest)
	if err != nil {
		return fmt.Errorf("metaregistrar: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
