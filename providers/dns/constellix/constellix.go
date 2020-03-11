// Package constellix implements a DNS provider for solving the DNS-01 challenge using Constellix DNS.
package constellix

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/providers/dns/constellix/internal"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	APIKey             string
	SecretKey          string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("CONSTELLIX_TTL", 300),
		PropagationTimeout: env.GetOrDefaultSecond("CONSTELLIX_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("CONSTELLIX_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("CONSTELLIX_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Constellix.
// Credentials must be passed in the environment variables:
// CONSTELLIX_API_KEY and CONSTELLIX_SECRET_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("CONSTELLIX_API_KEY", "CONSTELLIX_SECRET_KEY")
	if err != nil {
		return nil, fmt.Errorf("constellix: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["CONSTELLIX_API_KEY"]
	config.SecretKey = values["CONSTELLIX_SECRET_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Constellix.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("constellix: the configuration of the DNS provider is nil")
	}

	if config.SecretKey == "" || config.APIKey == "" {
		return nil, errors.New("constellix: incomplete credentials, missing secret key and/or API key")
	}

	tr, err := internal.NewTokenTransport(config.APIKey, config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("constellix: %w", err)
	}

	client := internal.NewClient(tr.Wrap(config.HTTPClient))

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("constellix: could not find zone for domain %q and fqdn %q : %w", domain, fqdn, err)
	}

	domainID, err := d.client.Domains.GetID(dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("constellix: failed to get domain ID: %w", err)
	}

	records, err := d.client.TxtRecords.GetAll(domainID)
	if err != nil {
		return fmt.Errorf("constellix: failed to get TXT records: %w", err)
	}

	recordName := getRecordName(fqdn, authZone)

	record := findRecords(records, recordName)

	// TXT record entry already existing
	if record != nil {
		if containsValue(record, value) {
			return nil
		}

		request := internal.RecordRequest{
			Name:       record.Name,
			TTL:        record.TTL,
			RoundRobin: append(record.RoundRobin, internal.RecordValue{Value: fmt.Sprintf(`"%s"`, value)}),
		}

		_, err = d.client.TxtRecords.Update(domainID, record.ID, request)
		if err != nil {
			return fmt.Errorf("constellix: failed to update TXT records: %w", err)
		}
		return nil
	}

	request := internal.RecordRequest{
		Name: recordName,
		TTL:  d.config.TTL,
		RoundRobin: []internal.RecordValue{
			{Value: fmt.Sprintf(`"%s"`, value)},
		},
	}

	_, err = d.client.TxtRecords.Create(domainID, request)
	if err != nil {
		return fmt.Errorf("constellix: failed to create TXT record %s: %w", fqdn, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("constellix: could not find zone for domain %q and fqdn %q : %w", domain, fqdn, err)
	}

	domainID, err := d.client.Domains.GetID(dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("constellix: failed to get domain ID: %w", err)
	}

	records, err := d.client.TxtRecords.GetAll(domainID)
	if err != nil {
		return fmt.Errorf("constellix: failed to get TXT records: %w", err)
	}

	recordName := getRecordName(fqdn, authZone)

	record := findRecords(records, recordName)
	if record == nil {
		return nil
	}

	if !containsValue(record, value) {
		return nil
	}

	// only 1 record value, the whole record must be deleted.
	if len(record.Value) == 1 {
		_, err = d.client.TxtRecords.Delete(domainID, record.ID)
		if err != nil {
			return fmt.Errorf("constellix: failed to delete TXT records: %w", err)
		}
		return nil
	}

	request := internal.RecordRequest{
		Name: record.Name,
		TTL:  record.TTL,
	}

	for _, val := range record.Value {
		if val.Value != fmt.Sprintf(`"%s"`, value) {
			request.RoundRobin = append(request.RoundRobin, val)
		}
	}

	_, err = d.client.TxtRecords.Update(domainID, record.ID, request)
	if err != nil {
		return fmt.Errorf("constellix: failed to update TXT records: %w", err)
	}

	return nil
}

func findRecords(records []internal.Record, name string) *internal.Record {
	for _, r := range records {
		if r.Name == name {
			return &r
		}
	}

	return nil
}

func containsValue(record *internal.Record, value string) bool {
	for _, val := range record.Value {
		if val.Value == fmt.Sprintf(`"%s"`, value) {
			return true
		}
	}

	return false
}

func getRecordName(fqdn, authZone string) string {
	return fqdn[0 : len(fqdn)-len(authZone)-1]
}
