// Package netcupv1 implements a DNS provider for solving the DNS-01 challenge using Netcup v1.
package netcupv1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/netcupv1/internal"
)

// Environment variables names.
const (
	envNamespace = "NETCUPV1_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 15*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 30*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for Netcup v1.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("netcupv1: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Netcup v1.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("netcupv1: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("netcupv1: %w", err)
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
		return fmt.Errorf("netcupv1: could not find zone for domain %q: %w", domain, err)
	}

	zone, err := d.findZone(ctx, info.Domain())
	if err != nil {
		return fmt.Errorf("netcupv1: %w", err)
	}

	subDomain, err := extractSubDomain(authZone, zone)
	if err != nil {
		return fmt.Errorf("netcupv1: %w", err)
	}

	_, err = d.client.AddChallenge(ctx, zone.ID, subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("netcupv1: add challenge: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("netcupv1: could not find zone for domain %q: %w", domain, err)
	}

	zone, err := d.findZone(ctx, info.Domain())
	if err != nil {
		return fmt.Errorf("netcupv1: %w", err)
	}

	subDomain, err := extractSubDomain(authZone, zone)
	if err != nil {
		return fmt.Errorf("netcupv1: %w", err)
	}

	err = d.client.DeleteChallenge(ctx, zone.ID, subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("netcupv1: delete challenge: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, domain string) (*internal.Domain, error) {
	for dom := range dns01.UnFqdnDomainsSeq(domain) {
		dr := &internal.DomainsRequest{
			FQDN: dom,
		}

		domains, err := d.client.GetDomains(ctx, dr)
		if err != nil {
			log.Debug("get domains", log.ErrorAttr(err))

			continue
		}

		if len(domains) == 0 {
			log.Debug("no domains found", log.DomainAttr(dom))

			continue
		}

		if len(domains) > 1 {
			log.Warn("multiple domains found", log.DomainAttr(dom))
		}

		return &domains[0], nil
	}

	return nil, fmt.Errorf("no domains found for %q", domain)
}

func extractSubDomain(authZone string, zone *internal.Domain) (string, error) {
	if dns01.UnFqdn(authZone) == zone.FQDN {
		return "@", nil
	}

	return dns01.ExtractSubDomain(authZone, zone.FQDN)
}
