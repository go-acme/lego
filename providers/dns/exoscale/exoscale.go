// Package exoscale implements a DNS provider for solving the DNS-01 challenge using exoscale DNS.
package exoscale

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/exoscale/egoscale"
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
)

const defaultBaseURL = "https://api.exoscale.com/dns"

// Environment variables names.
const (
	envNamespace = "EXOSCALE_"

	EnvAPISecret = envNamespace + "API_SECRET"
	EnvAPIKey    = envNamespace + "API_KEY"
	EnvEndpoint  = envNamespace + "ENDPOINT"

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
	HTTPClient         *http.Client
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *egoscale.Client
}

// NewDNSProvider Credentials must be passed in the environment variables:
// EXOSCALE_API_KEY, EXOSCALE_API_SECRET, EXOSCALE_ENDPOINT.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvAPIKey, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("exoscale: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.APIKey = values[EnvAPIKey]
	config.APISecret = values[EnvAPISecret]
	config.Endpoint = env.GetOrFile(conf, EnvEndpoint)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Exoscale.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" || config.APISecret == "" {
		return nil, errors.New("exoscale: credentials missing")
	}

	if config.Endpoint == "" {
		config.Endpoint = defaultBaseURL
	}

	client := egoscale.NewClient(config.Endpoint, config.APIKey, config.APISecret)
	client.HTTPClient = config.HTTPClient

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	zone, recordName, err := d.FindZoneAndRecordName(fqdn, domain)
	if err != nil {
		return err
	}

	recordID, err := d.FindExistingRecordID(zone, recordName)
	if err != nil {
		return err
	}

	if recordID == 0 {
		record := egoscale.DNSRecord{
			Name:       recordName,
			TTL:        d.config.TTL,
			Content:    value,
			RecordType: "TXT",
		}

		_, err := d.client.CreateRecord(ctx, zone, record)
		if err != nil {
			return errors.New("Error while creating DNS record: " + err.Error())
		}
	} else {
		record := egoscale.UpdateDNSRecord{
			ID:         recordID,
			Name:       recordName,
			TTL:        d.config.TTL,
			Content:    value,
			RecordType: "TXT",
		}

		_, err := d.client.UpdateRecord(ctx, zone, record)
		if err != nil {
			return errors.New("Error while updating DNS record: " + err.Error())
		}
	}

	return nil
}

// CleanUp removes the record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	zone, recordName, err := d.FindZoneAndRecordName(fqdn, domain)
	if err != nil {
		return err
	}

	recordID, err := d.FindExistingRecordID(zone, recordName)
	if err != nil {
		return err
	}

	if recordID != 0 {
		err = d.client.DeleteRecord(ctx, zone, recordID)
		if err != nil {
			return errors.New("Error while deleting DNS record: " + err.Error())
		}
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// FindExistingRecordID Query Exoscale to find an existing record for this name.
// Returns nil if no record could be found.
func (d *DNSProvider) FindExistingRecordID(zone, recordName string) (int64, error) {
	ctx := context.Background()
	records, err := d.client.GetRecords(ctx, zone)
	if err != nil {
		return -1, errors.New("Error while retrievening DNS records: " + err.Error())
	}
	for _, record := range records {
		if record.Name == recordName {
			return record.ID, nil
		}
	}
	return 0, nil
}

// FindZoneAndRecordName Extract DNS zone and DNS entry name.
func (d *DNSProvider) FindZoneAndRecordName(fqdn, domain string) (string, string, error) {
	zone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return "", "", err
	}
	zone = dns01.UnFqdn(zone)
	name := dns01.UnFqdn(fqdn)
	name = name[:len(name)-len("."+zone)]

	return zone, name, nil
}
