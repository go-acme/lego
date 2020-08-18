// Package dynu implements a DNS provider for solving the DNS-01 challenge using Dynu DNS.
package dynu

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/providers/dns/dynu/internal"
	"github.com/miekg/dns"
)

// Environment variables names.
const (
	envNamespace = "DYNU_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, 3*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, 10*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for Dynu.
// Credentials must be passed in the environment variables.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("dynu: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Dynu.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dynu: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("dynu: incomplete credentials, missing API key")
	}

	tr, err := internal.NewTokenTransport(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("dynu: %w", err)
	}

	client := internal.NewClient()
	client.HTTPClient = tr.Wrap(config.HTTPClient)

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	rootDomain, err := d.client.GetRootDomain(domain)
	if err != nil {
		return fmt.Errorf("dynu: could not find root domain for %s: %w", domain, err)
	}

	records, err := d.client.GetRecords(dns01.UnFqdn(fqdn), "TXT")
	if err != nil {
		return fmt.Errorf("dynu: failed to get records for %s: %w", domain, err)
	}

	for _, record := range records {
		// the record already exist
		if record.Hostname == dns01.UnFqdn(fqdn) && record.TextData == value {
			return nil
		}
	}

	record := internal.DNSRecord{
		Type:       "TXT",
		DomainName: rootDomain.DomainName,
		Hostname:   dns01.UnFqdn(fqdn),
		NodeName:   dns01.UnFqdn(strings.TrimSuffix(fqdn, dns.Fqdn(domain))),
		TextData:   value,
		State:      true,
		TTL:        d.config.TTL,
	}

	err = d.client.AddNewRecord(rootDomain.ID, record)
	if err != nil {
		return fmt.Errorf("dynu: failed to add record to %s: %w", domain, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	rootDomain, err := d.client.GetRootDomain(domain)
	if err != nil {
		return fmt.Errorf("dynu: could not find root domain for %s: %w", domain, err)
	}

	records, err := d.client.GetRecords(dns01.UnFqdn(fqdn), "TXT")
	if err != nil {
		return fmt.Errorf("dynu: failed to get records for %s: %w", domain, err)
	}

	for _, record := range records {
		if record.Hostname == dns01.UnFqdn(fqdn) && record.TextData == value {
			err = d.client.DeleteRecord(rootDomain.ID, record.ID)
			if err != nil {
				return fmt.Errorf("dynu: failed to remove TXT record for %s: %w", domain, err)
			}
		}
	}

	return nil
}
