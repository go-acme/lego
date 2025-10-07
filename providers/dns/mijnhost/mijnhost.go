// Package mijnhost implements a DNS provider for solving the DNS-01 challenge using mijn.host DNS.
package mijnhost

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/mijnhost/internal"
)

// Environment variables names.
const (
	envNamespace = "MIJNHOST_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const txtType = "TXT"

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, 5*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for mijn.host DNS.
// MIJNHOST_API_KEY must be passed in the environment variables.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("mijnhost: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for mijn.host DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("mijnhost: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("mijnhost: APIKey is missing")
	}

	client := internal.NewClient(config.APIKey)

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

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	domains, err := d.client.ListDomains(ctx)
	if err != nil {
		return fmt.Errorf("mijnhost: list domains: %w", err)
	}

	dom, err := findDomain(domains, domain)
	if err != nil {
		return fmt.Errorf("mijnhost: find domain: %w", err)
	}

	records, err := d.client.GetRecords(ctx, dom.Domain)
	if err != nil {
		return fmt.Errorf("mijnhost: get records: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, dom.Domain)
	if err != nil {
		return fmt.Errorf("mijnhost: %w", err)
	}

	record := internal.Record{
		Type:  txtType,
		Name:  subDomain,
		Value: info.Value,
		TTL:   d.config.TTL,
	}

	// mijn.host doesn't support multiple values for a domain,
	// so we removed existing record for the subdomain.
	cleanedRecords := filterRecords(records, func(record internal.Record) bool {
		return record.Type == txtType && (record.Name == subDomain || record.Name == dns01.UnFqdn(info.EffectiveFQDN))
	})

	cleanedRecords = append(cleanedRecords, record)

	err = d.client.UpdateRecords(ctx, dom.Domain, cleanedRecords)
	if err != nil {
		return fmt.Errorf("mijnhost: update records: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	domains, err := d.client.ListDomains(ctx)
	if err != nil {
		return fmt.Errorf("mijnhost: list domains: %w", err)
	}

	dom, err := findDomain(domains, domain)
	if err != nil {
		return fmt.Errorf("mijnhost: find domain: %w", err)
	}

	records, err := d.client.GetRecords(ctx, dom.Domain)
	if err != nil {
		return fmt.Errorf("mijnhost: get records: %w", err)
	}

	cleanedRecords := filterRecords(records, func(record internal.Record) bool {
		return record.Type == txtType && record.Value == info.Value
	})

	err = d.client.UpdateRecords(ctx, dom.Domain, cleanedRecords)
	if err != nil {
		return fmt.Errorf("mijnhost: update records: %w", err)
	}

	return nil
}

func findDomain(domains []internal.Domain, fqdn string) (internal.Domain, error) {
	for domain := range dns01.UnFqdnDomainsSeq(fqdn) {
		for _, dom := range domains {
			if dom.Domain == domain {
				return dom, nil
			}
		}
	}

	return internal.Domain{}, fmt.Errorf("domain %s not found", fqdn)
}

func filterRecords(records []internal.Record, fn func(record internal.Record) bool) []internal.Record {
	var newRecords []internal.Record

	for _, record := range records {
		if record.Type == "TXT" && fn(record) {
			continue
		}

		newRecords = append(newRecords, record)
	}

	return newRecords
}
