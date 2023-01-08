// Package pdns implements a DNS provider for solving the DNS-01 challenge using PowerDNS nameserver.
package pdns

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "PDNS_"

	EnvAPIKey = envNamespace + "API_KEY"
	EnvAPIURL = envNamespace + "API_URL"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvServerName         = envNamespace + "SERVER_NAME"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	Host               *url.URL
	ServerName         string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 2*time.Second),
		ServerName:         env.GetOrDefaultString(EnvServerName, "localhost"),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	apiVersion int
	config     *Config
}

// NewDNSProvider returns a DNSProvider instance configured for pdns.
// Credentials must be passed in the environment variable:
// PDNS_API_URL and PDNS_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvAPIURL)
	if err != nil {
		return nil, fmt.Errorf("pdns: %w", err)
	}

	hostURL, err := url.Parse(values[EnvAPIURL])
	if err != nil {
		return nil, fmt.Errorf("pdns: %w", err)
	}

	config := NewDefaultConfig()
	config.Host = hostURL
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for pdns.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("pdns: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("pdns: API key missing")
	}

	if config.Host == nil || config.Host.Host == "" {
		return nil, errors.New("pdns: API URL missing")
	}

	d := &DNSProvider{config: config}

	apiVersion, err := d.getAPIVersion()
	if err != nil {
		log.Warnf("pdns: failed to get API version %v", err)
	}
	d.apiVersion = apiVersion

	return d, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := d.getHostedZone(fqdn)
	if err != nil {
		return fmt.Errorf("pdns: %w", err)
	}

	name := fqdn

	// pre-v1 API wants non-fqdn
	if d.apiVersion == 0 {
		name = dns01.UnFqdn(fqdn)
	}

	rec := Record{
		Content:  "\"" + value + "\"",
		Disabled: false,

		// pre-v1 API
		Type: "TXT",
		Name: name,
		TTL:  d.config.TTL,
	}

	// Look for existing records.
	existingRrSet, err := d.findTxtRecord(fqdn)
	if err != nil {
		return fmt.Errorf("pdns: %w", err)
	}

	// merge the existing and new records
	var records []Record
	if existingRrSet != nil {
		records = existingRrSet.Records
	}
	records = append(records, rec)

	rrsets := rrSets{
		RRSets: []rrSet{
			{
				Name:       name,
				ChangeType: "REPLACE",
				Type:       "TXT",
				Kind:       "Master",
				TTL:        d.config.TTL,
				Records:    records,
			},
		},
	}

	body, err := json.Marshal(rrsets)
	if err != nil {
		return fmt.Errorf("pdns: %w", err)
	}

	_, err = d.sendRequest(http.MethodPatch, zone.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("pdns: %w", err)
	}

	return d.notify(zone)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zone, err := d.getHostedZone(fqdn)
	if err != nil {
		return fmt.Errorf("pdns: %w", err)
	}

	set, err := d.findTxtRecord(fqdn)
	if err != nil {
		return fmt.Errorf("pdns: %w", err)
	}
	if set == nil {
		return fmt.Errorf("pdns: no existing record found for %s", fqdn)
	}

	rrsets := rrSets{
		RRSets: []rrSet{
			{
				Name:       set.Name,
				Type:       set.Type,
				ChangeType: "DELETE",
			},
		},
	}
	body, err := json.Marshal(rrsets)
	if err != nil {
		return fmt.Errorf("pdns: %w", err)
	}

	_, err = d.sendRequest(http.MethodPatch, zone.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("pdns: %w", err)
	}

	return d.notify(zone)
}
