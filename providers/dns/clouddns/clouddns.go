// Package clouddns implements a DNS provider for solving the DNS-01 challenge using CloudDNS API.
package clouddns

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/clouddns/internal"
)

// Environment variables names.
const (
	envNamespace = "CLOUDDNS_"

	EnvClientID = envNamespace + "CLIENT_ID"
	EnvEmail    = envNamespace + "EMAIL"
	EnvPassword = envNamespace + "PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the DNSProvider.
type Config struct {
	ClientID string
	Email    string
	Password string

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 5*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *internal.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for CloudDNS.
// Credentials must be passed in the environment variables:
// CLOUDDNS_CLIENT_ID, CLOUDDNS_EMAIL, CLOUDDNS_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvClientID, EnvEmail, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("clouddns: %w", err)
	}

	config := NewDefaultConfig()
	config.ClientID = values[EnvClientID]
	config.Email = values[EnvEmail]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for CloudDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("clouddns: the configuration of the DNS provider is nil")
	}

	if config.ClientID == "" || config.Email == "" || config.Password == "" {
		return nil, errors.New("clouddns: credentials missing")
	}

	client := internal.NewClient(config.ClientID, config.Email, config.Password, config.TTL)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		client: client,
		config: config,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.CreateRecord(domain, token, fqdn, value)
}

// CreateRecord creates a TXT record to fulfill the DNS-01 challenge.
func (d *DNSProvider) CreateRecord(domain, token, fqdn, value string) error {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("clouddns: %w", err)
	}

	err = d.client.AddRecord(authZone, fqdn, value)
	if err != nil {
		return fmt.Errorf("clouddns: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.DeleteRecord(domain, token, fqdn, value)
}

// DeleteRecord removes a creates a TXT record from the provider.
func (d *DNSProvider) DeleteRecord(domain, token, fqdn, value string) error {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("clouddns: %w", err)
	}

	err = d.client.DeleteRecord(authZone, fqdn)
	if err != nil {
		return fmt.Errorf("clouddns: %w", err)
	}

	return nil
}
