// Package selectel implements a DNS provider for solving the DNS-01
// challenge using Selectel Domains API.
package selectel

// Selectel Domain API reference: https://my.selectel.ru/domains/doc
// Token: https://my.selectel.ru/profile/apikeys

import (
	"errors"
	"fmt"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

const (
	defaultBaseURL = "https://api.selectel.ru/domains/v1"
	minTTL         = 60

	envSelectelBaseURL            = "SELECTEL_BASE_URL"
	envSelectelAPIToken           = "SELECTEL_API_TOKEN"
	envSelectelTTL                = "SELECTEL_TTL"
	envSelectelPropagationTimeout = "SELECTEL_PROPAGATION_TIMEOUT"
	envSelectelPollingInterval    = "SELECTEL_POLLING_INTERVAL"
	envSelectelHTTPTimeout        = "SELECTEL_HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPTimeout        time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            env.GetOrDefaultString(envSelectelBaseURL, defaultBaseURL),
		TTL:                env.GetOrDefaultInt(envSelectelTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(envSelectelPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(envSelectelPollingInterval, 2*time.Second),
		HTTPTimeout:        env.GetOrDefaultSecond(envSelectelHTTPTimeout, 30*time.Second),
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	config *Config
	client *Client
}

// NewDNSProvider returns a DNSProvider instance configured for Selectel Domains API.
// API token must be passed in the environment variable SELECTEL_API_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(envSelectelAPIToken)
	if err != nil {
		return nil, fmt.Errorf("selectel: %v", err)
	}

	config := NewDefaultConfig()
	config.Token = values[envSelectelAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for selectel.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("selectel: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, fmt.Errorf("selectel: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("selectel: invalid TTL, TTL (%d) must be greater than %d",
			config.TTL,
			minTTL)
	}

	// Init client options
	opts := ClientOpts{
		BaseURL:   config.BaseURL,
		Token:     config.Token,
		UserAgent: "lego",
		Timeout:   config.HTTPTimeout,
	}

	// Init Selectel DNS client
	client := NewSelectelDNS(opts)

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the Timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill DNS-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	// Get domain via Selectel Domain API
	domainObj, err := d.client.GetDomainByName(domain)
	if err != nil {
		return err
	}

	// Set TXT record for DNS-01 challenge
	txtRecord := Record{
		Type:    "TXT",
		TTL:     d.config.TTL,
		Name:    fqdn,
		Content: value,
	}
	_, err = d.client.AddRecord(domainObj.ID, txtRecord)
	return err
}

// CleanUp removes a TXT record used for DNS-01 challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	// Get domain via Selectel Domain API
	domainObj, err := d.client.GetDomainByName(domain)
	if err != nil {
		return err
	}

	// Get list records for domain
	records, err := d.client.ListRecords(domainObj.ID)
	if err != nil {
		panic(err)
	}

	// Delete records with specific FQDN
	for _, record := range records {
		if record.Name == fqdn {
			err = d.client.DeleteRecord(domainObj.ID, record.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
