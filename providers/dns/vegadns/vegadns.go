// Package vegadns implements a DNS provider for solving the DNS-01 challenge using VegaDNS.
package vegadns

import (
	"errors"
	"fmt"
	"time"

	vegaClient "github.com/OpenDNS/vegadns2client"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "VEGADNS_"

	EnvKey    = "SECRET_VEGADNS_KEY"
	EnvSecret = "SECRET_VEGADNS_SECRET"
	EnvURL    = envNamespace + "URL"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	APIKey             string
	APISecret          string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 10),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 12*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 1*time.Minute),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client vegaClient.VegaDNSClient
}

// NewDNSProvider returns a DNSProvider instance configured for VegaDNS.
// Credentials must be passed in the environment variables:
// VEGADNS_URL, SECRET_VEGADNS_KEY, SECRET_VEGADNS_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvURL)
	if err != nil {
		return nil, fmt.Errorf("vegadns: %w", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values[EnvURL]
	config.APIKey = env.GetOrFile(EnvKey)
	config.APISecret = env.GetOrFile(EnvSecret)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for VegaDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("vegadns: the configuration of the DNS provider is nil")
	}

	vega := vegaClient.NewVegaDNSClient(config.BaseURL)
	vega.APIKey = config.APIKey
	vega.APISecret = config.APISecret

	return &DNSProvider{client: vega, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	_, domainID, err := d.client.GetAuthZone(fqdn)
	if err != nil {
		return fmt.Errorf("vegadns: can't find Authoritative Zone for %s in Present: %w", fqdn, err)
	}

	err = d.client.CreateTXT(domainID, fqdn, value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("vegadns: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	_, domainID, err := d.client.GetAuthZone(fqdn)
	if err != nil {
		return fmt.Errorf("vegadns: can't find Authoritative Zone for %s in CleanUp: %w", fqdn, err)
	}

	txt := dns01.UnFqdn(fqdn)

	recordID, err := d.client.GetRecordID(domainID, txt, "TXT")
	if err != nil {
		return fmt.Errorf("vegadns: couldn't get Record ID in CleanUp: %w", err)
	}

	err = d.client.DeleteRecord(recordID)
	if err != nil {
		return fmt.Errorf("vegadns: %w", err)
	}
	return nil
}
