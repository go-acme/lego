// Package opusdns implements a DNS provider for solving the DNS-01 challenge using OpusDNS.
package opusdns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/opusdns/opusdns-go-client/models"
	"github.com/opusdns/opusdns-go-client/opusdns"
)

// Environment variables names.
const (
	envNamespace = "OPUSDNS_"

	EnvAPIKey      = envNamespace + "API_KEY"
	EnvAPIEndpoint = envNamespace + "API_ENDPOINT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const defaultTTL = 60

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	APIEndpoint        string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 4*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *opusdns.Client

	// findZoneByFqdn determines the DNS zone of a FQDN.
	// It is overridable for testing purposes.
	findZoneByFqdn func(ctx context.Context, fqdn string) (string, error)
}

// NewDNSProvider returns a DNSProvider instance configured for OpusDNS.
// Credentials must be passed in the environment variable: OPUSDNS_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("opusdns: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.APIEndpoint = env.GetOrDefaultString(EnvAPIEndpoint, opusdns.DefaultAPIEndpoint)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OpusDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("opusdns: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("opusdns: incomplete credentials, missing API key")
	}

	opts := []opusdns.Option{
		opusdns.WithAPIKey(config.APIKey),
		opusdns.WithMaxRetries(0),
	}

	if config.APIEndpoint != "" {
		opts = append(opts, opusdns.WithAPIEndpoint(config.APIEndpoint))
	}

	if config.HTTPClient != nil {
		opts = append(opts, opusdns.WithHTTPClient(config.HTTPClient))
	}

	client, err := opusdns.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("opusdns: failed to create client: %w", err)
	}

	return &DNSProvider{config: config, client: client, findZoneByFqdn: dns01.DefaultClient().FindZoneByFqdn}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := d.findZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("opusdns: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("opusdns: %w", err)
	}

	zoneName := dns01.UnFqdn(authZone)

	err = d.client.DNS.UpsertRecord(ctx, zoneName, models.Record{
		Name:  subDomain,
		Type:  models.RRSetTypeTXT,
		TTL:   d.config.TTL,
		RData: info.Value,
	})
	if err != nil {
		return fmt.Errorf("opusdns: failed to create TXT record for %q: %w", info.EffectiveFQDN, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := d.findZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("opusdns: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("opusdns: %w", err)
	}

	zoneName := dns01.UnFqdn(authZone)

	err = d.client.DNS.DeleteRecord(ctx, zoneName, models.Record{
		Name:  subDomain,
		Type:  models.RRSetTypeTXT,
		TTL:   d.config.TTL,
		RData: info.Value,
	})
	if err != nil {
		return fmt.Errorf("opusdns: failed to delete TXT record for %q: %w", info.EffectiveFQDN, err)
	}

	return nil
}
