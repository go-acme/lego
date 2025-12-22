// Package ispconfigddns implements a DNS provider for solving the DNS-01 challenge using ISPConfig 3 Dynamic DNS (DDNS) Module.
package ispconfigddns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/ispconfigddns/internal"
)

// Environment variables names.
const (
	envNamespace = "ISPCONFIG_DDNS_"

	EnvServerURL = envNamespace + "SERVER_URL"
	EnvToken     = envNamespace + "TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ServerURL string
	Token     string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 3600),
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

// NewDNSProvider returns a DNSProvider instance configured for ISPConfig 3 Dynamic DNS (DDNS) Module.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvServerURL, EnvToken)
	if err != nil {
		return nil, fmt.Errorf("ispconfig (DDNS module): %w", err)
	}

	config := NewDefaultConfig()
	config.ServerURL = values[EnvServerURL]
	config.Token = values[EnvToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ISPConfig 3 Dynamic DNS (DDNS) Module.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ispconfig (DDNS module): the configuration of the DNS provider is nil")
	}

	if config.ServerURL == "" {
		return nil, errors.New("ispconfig (DDNS module): missing server URL")
	}

	if config.Token == "" {
		return nil, errors.New("ispconfig (DDNS module): missing token")
	}

	client, err := internal.NewClient(config.ServerURL, config.Token)
	if err != nil {
		return nil, fmt.Errorf("ispconfig (DDNS module): %w", err)
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

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to control checking compliance to spec.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ispconfig (DDNS module): could not find zone for domain %q: %w", domain, err)
	}

	err = d.client.AddTXTRecord(context.Background(), dns01.UnFqdn(zone), info.EffectiveFQDN, info.Value)
	if err != nil {
		return fmt.Errorf("ispconfig (DDNS module): add record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ispconfig (DDNS module): could not find zone for domain %q: %w", domain, err)
	}

	err = d.client.DeleteTXTRecord(context.Background(), dns01.UnFqdn(zone), info.EffectiveFQDN, info.Value)
	if err != nil {
		return fmt.Errorf("ispconfig (DDNS module): delete record: %w", err)
	}

	return nil
}
