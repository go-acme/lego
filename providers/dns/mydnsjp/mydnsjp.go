// Package mydnsjp implements a DNS provider for solving the DNS-01 challenge using MyDNS.jp.
package mydnsjp

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
	"github.com/go-acme/lego/v4/providers/dns/mydnsjp/internal"
)

// Environment variables names.
const (
	envNamespace = "MYDNSJP_"

	EnvMasterID = envNamespace + "MASTER_ID"
	EnvPassword = envNamespace + "PASSWORD"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	MasterID           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
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

// NewDNSProvider returns a DNSProvider instance configured for MyDNS.jp.
// Credentials must be passed in the environment variables: MYDNSJP_MASTER_ID and MYDNSJP_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvMasterID, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("mydnsjp: %w", err)
	}

	config := NewDefaultConfig()
	config.MasterID = values[EnvMasterID]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for MyDNS.jp.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("mydnsjp: the configuration of the DNS provider is nil")
	}

	if config.MasterID == "" || config.Password == "" {
		return nil, errors.New("mydnsjp: some credentials information are missing")
	}

	client := internal.NewClient(config.MasterID, config.Password)

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
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// TODO(ldez) replace domain by FQDN to follow CNAME.
	err := d.client.AddTXTRecord(context.Background(), domain, info.Value)
	if err != nil {
		return fmt.Errorf("mydnsjp: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// TODO(ldez) replace domain by FQDN to follow CNAME.
	err := d.client.DeleteTXTRecord(context.Background(), domain, info.Value)
	if err != nil {
		return fmt.Errorf("mydnsjp: %w", err)
	}
	return nil
}
