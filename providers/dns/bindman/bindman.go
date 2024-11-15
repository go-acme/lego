// Package bindman implements a DNS provider for solving the DNS-01 challenge.
package bindman

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/labbsr0x/bindman-dns-webhook/src/client"
)

// Environment variables names.
const (
	envNamespace = "BINDMAN_"

	EnvManagerAddress = envNamespace + "MANAGER_ADDRESS"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	BaseURL            string
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, time.Minute),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *client.DNSWebhookClient
}

// NewDNSProvider returns a DNSProvider instance configured for Bindman.
// BINDMAN_MANAGER_ADDRESS should have the scheme, hostname, and port (if required) of the authoritative Bindman Manager server.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvManagerAddress)
	if err != nil {
		return nil, fmt.Errorf("bindman: %w", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values[EnvManagerAddress]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Bindman.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("bindman: the configuration of the DNS provider is nil")
	}

	if config.BaseURL == "" {
		return nil, errors.New("bindman: bindman manager address missing")
	}

	bClient, err := client.New(config.BaseURL, config.HTTPClient)
	if err != nil {
		return nil, fmt.Errorf("bindman: %w", err)
	}

	return &DNSProvider{config: config, client: bClient}, nil
}

// Present creates a TXT record using the specified parameters.
// This will *not* create a subzone to contain the TXT record,
// so make sure the FQDN specified is within an extant zone.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	if err := d.client.AddRecord(info.EffectiveFQDN, "TXT", info.Value); err != nil {
		return fmt.Errorf("bindman: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	if err := d.client.RemoveRecord(info.EffectiveFQDN, "TXT"); err != nil {
		return fmt.Errorf("bindman: %w", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
