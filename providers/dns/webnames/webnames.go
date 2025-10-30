// Package webnames implements a DNS provider for solving the DNS-01 challenge using webnames.ru DNS.
package webnames

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/webnames/internal"
)

// Environment variables names.
const (
	envNamespace    = "WEBNAMESRU_"
	altEnvNamespace = "WEBNAMES_"

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
		PropagationTimeout: env.GetOneWithFallback(EnvPropagationTimeout, dns01.DefaultPropagationTimeout, env.ParseSecond, altEnvName(EnvPropagationTimeout)),
		PollingInterval:    env.GetOneWithFallback(EnvPollingInterval, dns01.DefaultPollingInterval, env.ParseSecond, altEnvName(EnvPollingInterval)),
		HTTPClient: &http.Client{
			Timeout: env.GetOneWithFallback(EnvHTTPTimeout, 20*time.Second, env.ParseSecond, altEnvName(EnvHTTPTimeout)),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a new DNS provider using
// environment variable WEBNAMESRU_API_KEY for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.GetWithFallback([]string{EnvAPIKey, altEnvName(EnvAPIKey)})
	if err != nil {
		return nil, fmt.Errorf("webnamesru: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Webnames.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("webnamesru: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("webnamesru: credentials missing")
	}

	client := internal.NewClient(config.APIKey)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("webnamesru: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("webnamesru: %w", err)
	}

	err = d.client.AddTXTRecord(context.Background(), dns01.UnFqdn(authZone), subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("webnamesru: failed to create TXT records [domain: %s, sub domain: %s]: %w",
			dns01.UnFqdn(authZone), subDomain, err)
	}

	return nil
}

// CleanUp clears Webnames TXT record.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("webnamesru: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("webnamesru: %w", err)
	}

	err = d.client.RemoveTXTRecord(context.Background(), dns01.UnFqdn(authZone), subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("webnamesru: failed to remove TXT records [domain: %s, sub domain: %s]: %w",
			dns01.UnFqdn(authZone), subDomain, err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func altEnvName(v string) string {
	return strings.ReplaceAll(v, envNamespace, altEnvNamespace)
}
