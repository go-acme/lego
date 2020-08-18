// Package cloudns implements a DNS provider for solving the DNS-01 challenge using ClouDNS DNS.
package cloudns

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/providers/dns/cloudns/internal"
)

// Environment variables names.
const (
	envNamespace = "CLOUDNS_"

	EnvAuthID       = envNamespace + "AUTH_ID"
	EnvSubAuthID    = envNamespace + "SUB_AUTH_ID"
	EnvAuthPassword = envNamespace + "AUTH_PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AuthID             string
	SubAuthID          string
	AuthPassword       string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, 4*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for ClouDNS.
// Credentials must be passed in the environment variables:
// CLOUDNS_AUTH_ID and CLOUDNS_AUTH_PASSWORD.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	var subAuthID string
	authID := env.GetOrFile(conf, EnvAuthID)
	if authID == "" {
		subAuthID = env.GetOrFile(conf, EnvSubAuthID)
	}

	if authID == "" && subAuthID == "" {
		return nil, fmt.Errorf("ClouDNS: some credentials information are missing: %s or %s", EnvAuthID, EnvSubAuthID)
	}

	values, err := env.Get(conf, EnvAuthPassword)
	if err != nil {
		return nil, fmt.Errorf("ClouDNS: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.AuthID = authID
	config.SubAuthID = subAuthID
	config.AuthPassword = values[EnvAuthPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ClouDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ClouDNS: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.AuthID, config.SubAuthID, config.AuthPassword)
	if err != nil {
		return nil, fmt.Errorf("ClouDNS: %w", err)
	}

	client.HTTPClient = config.HTTPClient

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := d.client.GetZone(fqdn)
	if err != nil {
		return fmt.Errorf("ClouDNS: %w", err)
	}

	err = d.client.AddTxtRecord(zone.Name, fqdn, value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("ClouDNS: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zone, err := d.client.GetZone(fqdn)
	if err != nil {
		return fmt.Errorf("ClouDNS: %w", err)
	}

	record, err := d.client.FindTxtRecord(zone.Name, fqdn)
	if err != nil {
		return fmt.Errorf("ClouDNS: %w", err)
	}

	if record == nil {
		return nil
	}

	err = d.client.RemoveTxtRecord(record.ID, zone.Name)
	if err != nil {
		return fmt.Errorf("ClouDNS: %w", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
