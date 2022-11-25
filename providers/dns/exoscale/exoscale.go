// Package exoscale implements a DNS provider for solving the DNS-01 challenge using Exoscale DNS.
package exoscale

import (
	"context"
	"errors"
	"fmt"
	"time"

	egoscale "github.com/exoscale/egoscale/v2"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Default Exoscale API endpoint.
const defaultBaseURL = "https://api.exoscale.com/v2"

// Default Exosacle API zone.
// Each data center location hosts the API and API zone determines which one to connect to.
const defaultAPIZone = "ch-gva-2"

// Environment variables names.
const (
	envNamespace = "EXOSCALE_"

	EnvAPISecret = envNamespace + "API_SECRET"
	EnvAPIKey    = envNamespace + "API_KEY"
	EnvEndpoint  = envNamespace + "ENDPOINT"
	EnvAPIZone   = envNamespace + "API_ZONE"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	APISecret          string
	Endpoint           string
	HTTPTimeout        time.Duration
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int64
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                int64(env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL)),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 60*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config  *Config
	client  *egoscale.Client
	apiZone string
}

// NewDNSProvider Credentials must be passed in the environment variables:
// EXOSCALE_API_KEY, EXOSCALE_API_SECRET, EXOSCALE_ENDPOINT.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("exoscale: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.APISecret = values[EnvAPISecret]
	config.Endpoint = env.GetOrFile(EnvEndpoint)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Exoscale.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("exoscale: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" || config.APISecret == "" {
		return nil, errors.New("exoscale: credentials missing")
	}

	if config.Endpoint == "" {
		config.Endpoint = defaultBaseURL
	}

	client, err := egoscale.NewClient(
		config.APIKey,
		config.APISecret,
		egoscale.ClientOptWithAPIEndpoint(config.Endpoint),
		egoscale.ClientOptWithTimeout(config.HTTPTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("exoscale: initializing client: %w", err)
	}

	return &DNSProvider{
		client:  client,
		config:  config,
		apiZone: env.GetOrDefaultString(EnvAPIZone, defaultAPIZone),
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zoneName, recordName, err := d.findZoneAndRecordName(fqdn)
	if err != nil {
		return err
	}

	zone, err := d.findExistingZone(zoneName)
	if err != nil {
		return fmt.Errorf("exoscale: %w", err)
	}
	if zone == nil {
		return fmt.Errorf("exoscale: zone %q not found", zoneName)
	}

	recordID, err := d.findExistingRecordID(*zone.ID, recordName)
	if err != nil {
		return fmt.Errorf("exoscale: %w", err)
	}

	recordType := "TXT"

	if recordID == "" {
		record := egoscale.DNSDomainRecord{
			Name:    &recordName,
			TTL:     &d.config.TTL,
			Content: &value,
			Type:    &recordType,
		}

		_, err = d.client.CreateDNSDomainRecord(ctx, d.apiZone, *zone.ID, &record)
		if err != nil {
			return fmt.Errorf("exoscale: error while creating DNS record: %w", err)
		}

		return nil
	}

	record := egoscale.DNSDomainRecord{
		ID:      &recordID,
		Name:    &recordName,
		TTL:     &d.config.TTL,
		Content: &value,
		Type:    &recordType,
	}

	err = d.client.UpdateDNSDomainRecord(ctx, d.apiZone, *zone.ID, &record)
	if err != nil {
		return fmt.Errorf("exoscale: error while updating DNS record: %w", err)
	}

	return nil
}

// CleanUp removes the record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zoneName, recordName, err := d.findZoneAndRecordName(fqdn)
	if err != nil {
		return err
	}

	zone, err := d.findExistingZone(zoneName)
	if err != nil {
		return fmt.Errorf("exoscale: %w", err)
	}
	if zone == nil {
		return fmt.Errorf("exoscale: zone %q not found", zoneName)
	}

	recordID, err := d.findExistingRecordID(*zone.ID, recordName)
	if err != nil {
		return err
	}

	if recordID != "" {
		err = d.client.DeleteDNSDomainRecord(ctx, d.apiZone, *zone.ID, &egoscale.DNSDomainRecord{ID: &recordID})
		if err != nil {
			return fmt.Errorf("exoscale: error while deleting DNS record: %w", err)
		}
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// findExistingZone Query Exoscale to find an existing zone for this name.
// Returns nil result if no zone could be found.
func (d *DNSProvider) findExistingZone(zoneName string) (*egoscale.DNSDomain, error) {
	ctx := context.Background()

	zones, err := d.client.ListDNSDomains(ctx, d.apiZone)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving DNS zones: %w", err)
	}

	for _, zone := range zones {
		if zone.UnicodeName != nil && *zone.UnicodeName == zoneName {
			return &zone, nil
		}
	}

	return nil, nil
}

// findExistingRecordID Query Exoscale to find an existing record for this name.
// Returns empty result if no record could be found.
func (d *DNSProvider) findExistingRecordID(zoneID, recordName string) (string, error) {
	ctx := context.Background()

	records, err := d.client.ListDNSDomainRecords(ctx, d.apiZone, zoneID)
	if err != nil {
		return "", fmt.Errorf("error while retrieving DNS records: %w", err)
	}

	recordType := "TXT"
	for _, record := range records {
		if record.Name != nil && *record.Name == recordName &&
			record.Type != nil && *record.Type == recordType {
			return *record.ID, nil
		}
	}

	return "", nil
}

// findZoneAndRecordName Extract DNS zone and DNS entry name.
func (d *DNSProvider) findZoneAndRecordName(fqdn string) (string, string, error) {
	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", "", err
	}

	zone = dns01.UnFqdn(zone)

	name := dns01.UnFqdn(fqdn)
	name = name[:len(name)-len("."+zone)]

	return zone, name, nil
}
