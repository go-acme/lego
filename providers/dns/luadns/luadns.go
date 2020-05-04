// Package luadns implements a DNS provider for solving the DNS-01 challenge using LuaDNS.
package luadns

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
)

const (
	// defaultBaseURL represents the API endpoint to call.
	defaultBaseURL = "https://api.luadns.com"
	minTTL         = 60
)

// Environment variables names.
const (
	envNamespace = "LUADNS_"

	EnvAPIUsername = envNamespace + "API_USERNAME"
	EnvAPIToken    = envNamespace + "API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIUsername        string
	APIToken           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the challenge.Provider interface.
type DNSProvider struct {
	config           *Config
	createdRecords   map[string]*DNSRecord
	createdRecordsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for LuaDNS.
// Credentials must be passed in the environment variables:
// LUADNS_API_USERNAME and LUADNS_API_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIUsername, EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("luadns: %w", err)
	}

	config := NewDefaultConfig()
	config.APIUsername = values[EnvAPIUsername]
	config.APIToken = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for LuaDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("luadns: the configuration of the DNS provider is nil")
	}

	if config.APIUsername == "" || config.APIToken == "" {
		return nil, errors.New("luadns: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("luadns: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	return &DNSProvider{config: config, createdRecords: make(map[string]*DNSRecord)}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("luadns: %w", err)
	}

	newRecord := NewDNSRecord{
		Name:    fqdn,
		Type:    "TXT",
		Content: value,
		TTL:     d.config.TTL,
	}

	record, err := d.createRecord(*zone, newRecord)
	if err != nil {
		return fmt.Errorf("luadns: API call failed: %w", err)
	}

	d.createdRecordsMu.Lock()
	d.createdRecords[token] = record
	d.createdRecordsMu.Unlock()

	return nil
}

// CleanUp deletes the TXT record created by Present.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	d.createdRecordsMu.Lock()
	record, ok := d.createdRecords[token]
	d.createdRecordsMu.Unlock()
	if !ok {
		return fmt.Errorf("luadns: unknown record ID for '%s'", fqdn)
	}

	err := d.deleteRecord(record)
	if err != nil {
		return fmt.Errorf("luadns: error deleting record: %w", err)
	}

	// Delete record from map
	d.createdRecordsMu.Lock()
	delete(d.createdRecords, token)
	d.createdRecordsMu.Unlock()

	return nil
}

func (d *DNSProvider) getHostedZone(domain string) (*DNSZone, error) {
	zones, err := d.listZones(domain)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	var zone *DNSZone = nil
	for i, z := range zones {
		if strings.HasSuffix(domain, z.Name) {
			if zone == nil || len(z.Name) > len(zone.Name) {
				zone = &zones[i]
			}
		}
	}
	if zone == nil {
		return nil, fmt.Errorf("no matching LuaDNS zone found for domain %s", domain)
	}

	return zone, nil
}
