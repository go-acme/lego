// Package hetzner implements a DNS provider for solving the DNS-01 challenge using Hetzner DNS.
package hetzner

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/hetzner/internal/hetznerv1"
	"github.com/go-acme/lego/v4/providers/dns/hetzner/internal/legacy"
)

// Environment variables names.
const (
	// Deprecated: use EnvAPIToken instead.
	EnvAPIKey   = legacy.EnvAPIKey
	EnvAPIToken = hetznerv1.EnvAPIToken

	EnvTTL                = hetznerv1.EnvTTL
	EnvPropagationTimeout = hetznerv1.EnvPropagationTimeout
	EnvPollingInterval    = hetznerv1.EnvPollingInterval
	EnvHTTPTimeout        = hetznerv1.EnvHTTPTimeout
)

const minTTL = 60

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	// Deprecated: use APIToken instead
	APIKey string

	APIToken string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	provider challenge.ProviderTimeout
}

// NewDNSProvider returns a DNSProvider instance configured for hetzner.
// Credentials must be passed in the environment variable: HETZNER_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	_, foundAPIToken := os.LookupEnv(EnvAPIToken)
	_, foundAPIKey := os.LookupEnv(EnvAPIKey)

	switch {
	case foundAPIToken:
		provider, err := hetznerv1.NewDNSProvider()
		if err != nil {
			return nil, err
		}

		return &DNSProvider{provider: provider}, nil

	case foundAPIKey:
		log.Warnf("APIKey (legacy Hetzner DNS API) is deprecated, please use APIToken (Hetzner Cloud API) instead.")

		provider, err := legacy.NewDNSProvider()
		if err != nil {
			return nil, err
		}

		return &DNSProvider{provider: provider}, nil

	default:
		provider, err := hetznerv1.NewDNSProvider()
		if err != nil {
			return nil, err
		}

		return &DNSProvider{provider: provider}, nil
	}
}

// NewDNSProviderConfig return a DNSProvider instance configured for hetzner.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("hetzner: the configuration of the DNS provider is nil")
	}

	switch {
	case config.APIToken != "":
		cfg := &hetznerv1.Config{
			APIToken:           config.APIToken,
			PropagationTimeout: config.PropagationTimeout,
			PollingInterval:    config.PollingInterval,
			TTL:                config.TTL,
			HTTPClient:         config.HTTPClient,
		}

		provider, err := hetznerv1.NewDNSProviderConfig(cfg)
		if err != nil {
			return nil, err
		}

		return &DNSProvider{provider: provider}, nil

	case config.APIKey != "":
		log.Warnf("%s (legacy Hetzner DNS API) is deprecated, please use %s (Hetzner Cloud API) instead.", EnvAPIKey, EnvAPIToken)

		cfg := &legacy.Config{
			APIKey:             config.APIKey,
			PropagationTimeout: config.PropagationTimeout,
			PollingInterval:    config.PollingInterval,
			TTL:                config.TTL,
			HTTPClient:         config.HTTPClient,
		}

		provider, err := legacy.NewDNSProviderConfig(cfg)
		if err != nil {
			return nil, err
		}

		return &DNSProvider{provider: provider}, nil
	}

	return nil, errors.New("hetzner: credentials missing")
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.provider.Timeout()
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	return d.provider.Present(domain, token, keyAuth)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	return d.provider.CleanUp(domain, token, keyAuth)
}
