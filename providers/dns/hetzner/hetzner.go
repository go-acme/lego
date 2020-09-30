// Package hetzner implements a DNS provider for solving the DNS-01 challenge using Hetzner DNS.
package hetzner

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/hetzner/internal"
)

const minTTL = 600

// Environment variables names.
const (
	envNamespace = "HETZNER_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
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

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for hetzner.
// Credentials must be passed in the environment variable: HETZNER_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("hetzner: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for hetzner.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("hetzner: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("hetzner: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("hetzner: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := internal.NewClient(config.APIKey)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the DNS-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.CreateRecord(domain, token, fqdn, value)
}

// Present creates a TXT record to fulfill the DNS-01 challenge.
func (d *DNSProvider) CreateRecord(domain, token, fqdn, value string) error {
	zone, err := getZone(fqdn)
	if err != nil {
		return fmt.Errorf("hetzner: failed to find zone: fqdn=%s: %w", fqdn, err)
	}

	zoneID, err := d.client.GetZoneID(zone)
	if err != nil {
		return fmt.Errorf("hetzner: %w", err)
	}

	record := internal.DNSRecord{
		Type:   "TXT",
		Name:   extractRecordName(fqdn, zone),
		Value:  value,
		TTL:    d.config.TTL,
		ZoneID: zoneID,
	}

	if err := d.client.CreateRecord(record); err != nil {
		return fmt.Errorf("hetzner: failed to add TXT record: fqdn=%s, zoneID=%s: %w", fqdn, zoneID, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.DeleteRecord(domain, token, fqdn, value)
}

// DeleteRecord removes the record matching the specified parameters.
func (d *DNSProvider) DeleteRecord(domain, token, fqdn, value string) error {
	zone, err := getZone(fqdn)
	if err != nil {
		return fmt.Errorf("hetzner: failed to find zone: fqdn=%s: %w", fqdn, err)
	}

	zoneID, err := d.client.GetZoneID(zone)
	if err != nil {
		return fmt.Errorf("hetzner: %w", err)
	}

	recordName := extractRecordName(fqdn, zone)

	record, err := d.client.GetTxtRecord(recordName, value, zoneID)
	if err != nil {
		return fmt.Errorf("hetzner: %w", err)
	}

	if err := d.client.DeleteRecord(record.ID); err != nil {
		return fmt.Errorf("hetzner: failed to delate TXT record: id=%s, name=%s: %w", record.ID, record.Name, err)
	}

	return nil
}

func extractRecordName(fqdn, zone string) string {
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+zone); idx != -1 {
		return name[:idx]
	}
	return name
}

func getZone(fqdn string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", err
	}

	return dns01.UnFqdn(authZone), nil
}
