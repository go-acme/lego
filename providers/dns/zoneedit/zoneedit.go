// Package zoneedit implements a DNS provider for solving the DNS-01 challenge using ZoneEdit.
package zoneedit

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/zoneedit/internal"
)

// Environment variables names.
const (
	envNamespace = "ZONEEDIT_"

	EnvUser     = envNamespace + "USER"
	EnAuthToken = envNamespace + "AUTH_TOKEN"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	User      string
	AuthToken string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
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

// NewDNSProvider returns a DNSProvider instance configured for ZoneEdit.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUser, EnAuthToken)
	if err != nil {
		return nil, fmt.Errorf("zoneedit: %w", err)
	}

	config := NewDefaultConfig()
	config.User = values[EnvUser]
	config.AuthToken = values[EnAuthToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ZoneEdit.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("zoneedit: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.User, config.AuthToken)
	if err != nil {
		return nil, fmt.Errorf("zoneedit: %w", err)
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

	err := d.client.CreateTXTRecord(dns01.UnFqdn(info.EffectiveFQDN), info.Value)
	if err != nil {
		return fmt.Errorf("zoneedit: create TXT record: %w", err)
	}

	// ERROR CODE="702" TEXT="Minimum 10 seconds between requests"
	time.Sleep(11 * time.Second)

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	err := d.client.DeleteTXTRecord(dns01.UnFqdn(info.EffectiveFQDN), info.Value)
	if err != nil {
		return fmt.Errorf("zoneedit: delete TXT record: %w", err)
	}

	// ERROR CODE="702" TEXT="Minimum 10 seconds between requests"
	time.Sleep(11 * time.Second)

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
