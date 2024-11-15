// Package rackspace implements a DNS provider for solving the DNS-01 challenge using rackspace DNS.
package rackspace

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/rackspace/internal"
)

// Environment variables names.
const (
	envNamespace = "RACKSPACE_"

	EnvUser   = envNamespace + "USER"
	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	APIUser            string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            internal.DefaultIdentityURL,
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
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

	token            string
	cloudDNSEndpoint string
}

// NewDNSProvider returns a DNSProvider instance configured for Rackspace.
// Credentials must be passed in the environment variables:
// RACKSPACE_USER and RACKSPACE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUser, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("rackspace: %w", err)
	}

	config := NewDefaultConfig()
	config.APIUser = values[EnvUser]
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Rackspace.
// It authenticates against the API, also grabbing the DNS Endpoint.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("rackspace: the configuration of the DNS provider is nil")
	}

	if config.APIUser == "" || config.APIKey == "" {
		return nil, errors.New("rackspace: credentials missing")
	}

	identifier := internal.NewIdentifier(config.HTTPClient, config.BaseURL)

	identity, err := identifier.Login(context.Background(), config.APIUser, config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("rackspace: %w", err)
	}

	// Iterate through the Service Catalog to get the DNS Endpoint
	var dnsEndpoint string
	for _, service := range identity.Access.ServiceCatalog {
		if service.Name == "cloudDNS" {
			dnsEndpoint = service.Endpoints[0].PublicURL
			break
		}
	}

	if dnsEndpoint == "" {
		return nil, errors.New("rackspace: failed to populate DNS endpoint, check Rackspace API for changes")
	}

	client, err := internal.NewClient(dnsEndpoint, identity.Access.Token.ID)
	if err != nil {
		return nil, fmt.Errorf("rackspace: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config:           config,
		client:           client,
		token:            identity.Access.Token.ID,
		cloudDNSEndpoint: dnsEndpoint,
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zoneID, err := d.client.GetHostedZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("rackspace: %w", err)
	}

	record := internal.Record{
		Name: dns01.UnFqdn(info.EffectiveFQDN),
		Type: "TXT",
		Data: info.Value,
		TTL:  d.config.TTL,
	}

	err = d.client.AddRecord(ctx, zoneID, record)
	if err != nil {
		return fmt.Errorf("rackspace: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zoneID, err := d.client.GetHostedZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("rackspace: %w", err)
	}

	record, err := d.client.FindTxtRecord(ctx, info.EffectiveFQDN, zoneID)
	if err != nil {
		return fmt.Errorf("rackspace: %w", err)
	}

	err = d.client.DeleteRecord(ctx, zoneID, record.ID)
	if err != nil {
		return fmt.Errorf("rackspace: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
