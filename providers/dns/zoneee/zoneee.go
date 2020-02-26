// Package zoneee implements a DNS provider for solving the DNS-01 challenge through zone.ee.
package zoneee

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Endpoint           *url.URL
	Username           string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	endpoint, _ := url.Parse(defaultEndpoint)

	return &Config{
		Endpoint: endpoint,
		// zone.ee can take up to 5min to propagate according to the support
		PropagationTimeout: env.GetOrDefaultSecond("ZONEEE_PROPAGATION_TIMEOUT", 5*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("ZONEEE_POLLING_INTERVAL", 5*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("ZONEEE_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider describes a provider for acme-proxy
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("ZONEEE_API_USER", "ZONEEE_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("zoneee: %w", err)
	}

	rawEndpoint := env.GetOrDefaultString("ZONEEE_ENDPOINT", defaultEndpoint)
	endpoint, err := url.Parse(rawEndpoint)
	if err != nil {
		return nil, fmt.Errorf("zoneee: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values["ZONEEE_API_USER"]
	config.APIKey = values["ZONEEE_API_KEY"]
	config.Endpoint = endpoint

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Zone.ee
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("zoneee: the configuration of the DNS provider is nil")
	}

	if config.Username == "" {
		return nil, errors.New("zoneee: credentials missing: username")
	}

	if config.APIKey == "" {
		return nil, errors.New("zoneee: credentials missing: API key")
	}

	if config.Endpoint == nil {
		return nil, errors.New("zoneee: the endpoint is missing")
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	record := txtRecord{
		Name:        fqdn[:len(fqdn)-1],
		Destination: value,
	}

	authZone, err := getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("zoneee: %w", err)
	}

	_, err = d.addTxtRecord(authZone, record)
	if err != nil {
		return fmt.Errorf("zoneee: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record previously created
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	_, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("zoneee: %w", err)
	}

	records, err := d.getTxtRecords(authZone)
	if err != nil {
		return fmt.Errorf("zoneee: %w", err)
	}

	var id string
	for _, record := range records {
		if record.Destination == value {
			id = record.ID
		}
	}

	if id == "" {
		return fmt.Errorf("zoneee: txt record does not exist for %s", value)
	}

	if err = d.removeTxtRecord(authZone, id); err != nil {
		return fmt.Errorf("zoneee: %w", err)
	}

	return nil
}

func getHostedZone(domain string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return "", err
	}

	zoneName := dns01.UnFqdn(authZone)
	return zoneName, nil
}
