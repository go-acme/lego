// Package ipv64 implements a DNS provider for solving the DNS-01 challenge using IPv64.
// See https://ipv64.net/healthcheck_updater_api for more info on updating TXT records.
package ipv64

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/ipv64/internal"
	"github.com/miekg/dns"
)

// Environment variables names.
const (
	envNamespace = "IPV64_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
	SequenceInterval   time.Duration // Deprecated: unused, will be removed in v5.
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
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

// NewDNSProvider returns a new DNS provider using
// environment variable IPV64_TOKEN for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("ipv64: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for IPv64.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ipv64: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("ipv64: credentials missing")
	}

	client := internal.NewClient(internal.OAuthStaticAccessToken(config.HTTPClient, config.APIKey))

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	sub, root, err := splitDomain(dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("ipv64: %w", err)
	}

	err = d.client.AddRecord(context.Background(), root, sub, "TXT", info.Value)
	if err != nil {
		return fmt.Errorf("ipv64: %w", err)
	}

	return nil
}

// CleanUp clears IPv64 TXT record.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	sub, root, err := splitDomain(dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("ipv64: %w", err)
	}

	err = d.client.DeleteRecord(context.Background(), root, sub, "TXT", info.Value)
	if err != nil {
		return fmt.Errorf("ipv64: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func splitDomain(full string) (string, string, error) {
	split := dns.Split(full)
	if len(split) < 3 {
		return "", "", fmt.Errorf("unsupported domain: %s", full)
	}

	if len(split) == 3 {
		return "", full, nil
	}

	domain := full[split[len(split)-3]:]
	subDomain := full[:split[len(split)-3]-1]

	return subDomain, domain, nil
}
