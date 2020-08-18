// Package clouddns implements a DNS provider for solving the DNS-01 challenge using CloudDNS API.
package clouddns

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/providers/dns/clouddns/internal"
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
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, 5*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 30*time.Second),
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
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvClientID, EnvEmail, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("clouddns: %w", err)
	}

	config := NewDefaultConfig(conf)
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
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

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
