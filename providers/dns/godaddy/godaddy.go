// Package godaddy implements a DNS provider for solving the DNS-01 challenge using godaddy DNS.
package godaddy

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/godaddy/internal"
)

const minTTL = 600

// Environment variables names.
const (
	envNamespace = "GODADDY_"

	EnvAPIKey    = envNamespace + "API_KEY"
	EnvAPISecret = envNamespace + "API_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	APISecret          string
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

// NewDNSProvider returns a DNSProvider instance configured for godaddy.
// Credentials must be passed in the environment variables:
// GODADDY_API_KEY and GODADDY_API_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("godaddy: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.APISecret = values[EnvAPISecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for godaddy.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("godaddy: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" || config.APISecret == "" {
		return nil, errors.New("godaddy: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("godaddy: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := internal.NewClient(config.APIKey, config.APISecret)

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

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	domainZone, err := getZone(fqdn)
	if err != nil {
		return fmt.Errorf("godaddy: failed to get zone: %w", err)
	}

	recordName := extractRecordName(fqdn, domainZone)

	records, err := d.client.GetRecords(domainZone, "TXT", recordName)
	if err != nil {
		return fmt.Errorf("godaddy: failed to get TXT records: %w", err)
	}

	var newRecords []internal.DNSRecord
	for _, record := range records {
		if record.Data != "" {
			newRecords = append(newRecords, record)
		}
	}

	record := internal.DNSRecord{
		Type: "TXT",
		Name: recordName,
		Data: value,
		TTL:  d.config.TTL,
	}
	newRecords = append(newRecords, record)

	err = d.client.UpdateTxtRecords(newRecords, domainZone, recordName)
	if err != nil {
		return fmt.Errorf("godaddy: failed to add TXT record: %w", err)
	}

	return nil
}

// CleanUp removes the record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	domainZone, err := getZone(fqdn)
	if err != nil {
		return fmt.Errorf("godaddy: failed to get zone: %w", err)
	}

	recordName := extractRecordName(fqdn, domainZone)

	records, err := d.client.GetRecords(domainZone, "TXT", recordName)
	if err != nil {
		return fmt.Errorf("godaddy: failed to get TXT records: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	allTxtRecords, err := d.client.GetRecords(domainZone, "TXT", "")
	if err != nil {
		return fmt.Errorf("godaddy: failed to get all TXT records: %w", err)
	}

	var recordsKeep []internal.DNSRecord
	for _, record := range allTxtRecords {
		if record.Data != value && record.Data != "" {
			recordsKeep = append(recordsKeep, record)
		}
	}

	// GoDaddy API don't provide a way to delete a record, an "empty" record must be added.
	if len(recordsKeep) == 0 {
		emptyRecord := internal.DNSRecord{Name: "empty", Data: ""}
		recordsKeep = append(recordsKeep, emptyRecord)
	}

	err = d.client.UpdateTxtRecords(recordsKeep, domainZone, "")
	if err != nil {
		return fmt.Errorf("godaddy: failed to remove TXT record: %w", err)
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
