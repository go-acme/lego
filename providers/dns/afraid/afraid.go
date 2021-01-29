// Package afraid implements a DNS provider for solving the DNS-01 challenge using Afraid freedns.
package afraid

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/afraid/internal"
)

// Environment variables names.
const (
	envNamespace = "AFRAID_"

	EnvLogin              = envNamespace + "LOGIN"
	EnvPassword           = envNamespace + "PASSWORD"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config Provider configuration.
type Config struct {
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 4*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	Client *internal.Client
}

// NewDNSProvider returns a new DNS provider which runs the program in the
// environment variable EXEC_PATH for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvLogin)
	if err != nil {
		return nil, fmt.Errorf("afraid: %w", err)
	}
	login := values[EnvLogin]

	values, err = env.Get(EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("afraid: %w", err)
	}
	pass := values[EnvPassword]

	client, err := internal.NewClient(login, pass)
	if err != nil {
		return nil, fmt.Errorf("afraid: %w", err)
	}
	client.SetHTTPTimeout(env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second))

	config := NewDefaultConfig()

	provider, err := NewDNSProviderConfig(config)
	if err != nil {
		return nil, fmt.Errorf("afraid: %w", err)
	}
	provider.Client = client
	return provider, nil
}

// NewDNSProviderConfig returns a new DNS provider which runs the given configuration
// for adding and removing the DNS record.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration is nil")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	err := d.Client.CreateTxtRecord(fqdn, value)
	if err != nil {
		return err
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	err := d.Client.DeleteTxtRecord(fqdn)
	if err != nil {
		return err
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.PropagationTimeout
}
