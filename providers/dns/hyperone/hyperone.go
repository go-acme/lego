// Package hyperone implements a DNS provider for solving the DNS-01 challenge using HyperOne.
package hyperone

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/hyperone/internal"
)

// Environment variables names.
const (
	envNamespace = "HYPERONE_"

	EnvPassportLocation = envNamespace + "PASSPORT_LOCATION"
	EnvAPIUrl           = envNamespace + "API_URL"
	EnvLocationID       = envNamespace + "LOCATION_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIEndpoint      string
	LocationID       string
	PassportLocation string

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *internal.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for HyperOne.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	config.PassportLocation = env.GetOrFile(EnvPassportLocation)
	config.LocationID = env.GetOrFile(EnvLocationID)
	config.APIEndpoint = env.GetOrFile(EnvAPIUrl)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for HyperOne.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config.PassportLocation == "" {
		var err error
		config.PassportLocation, err = GetDefaultPassportLocation()
		if err != nil {
			return nil, fmt.Errorf("hyperone: %w", err)
		}
	}

	passport, err := internal.LoadPassportFile(config.PassportLocation)
	if err != nil {
		return nil, fmt.Errorf("hyperone: %w", err)
	}

	client, err := internal.NewClient(config.APIEndpoint, config.LocationID, passport)
	if err != nil {
		return nil, fmt.Errorf("hyperone: failed to create client: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{client: client, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getHostedZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hyperone: failed to get zone for fqdn=%s: %w", info.EffectiveFQDN, err)
	}

	recordset, err := d.client.FindRecordset(zone.ID, "TXT", info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hyperone: fqdn=%s, zone ID=%s: %w", info.EffectiveFQDN, zone.ID, err)
	}

	if recordset == nil {
		_, err = d.client.CreateRecordset(zone.ID, "TXT", info.EffectiveFQDN, info.Value, d.config.TTL)
		if err != nil {
			return fmt.Errorf("hyperone: failed to create recordset: fqdn=%s, zone ID=%s, value=%s: %w", info.EffectiveFQDN, zone.ID, info.Value, err)
		}

		return nil
	}

	_, err = d.client.CreateRecord(zone.ID, recordset.ID, info.Value)
	if err != nil {
		return fmt.Errorf("hyperone: failed to create record: fqdn=%s, zone ID=%s, recordset ID=%s: %w", info.EffectiveFQDN, zone.ID, recordset.ID, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters and recordset if no other records are remaining.
// There is a small possibility that race will cause to delete recordset with records for other DNS Challenges.
func (d *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getHostedZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hyperone: failed to get zone for fqdn=%s: %w", info.EffectiveFQDN, err)
	}

	recordset, err := d.client.FindRecordset(zone.ID, "TXT", info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hyperone: fqdn=%s, zone ID=%s: %w", info.EffectiveFQDN, zone.ID, err)
	}

	if recordset == nil {
		return fmt.Errorf("hyperone: recordset to remove not found: fqdn=%s", info.EffectiveFQDN)
	}

	records, err := d.client.GetRecords(zone.ID, recordset.ID)
	if err != nil {
		return fmt.Errorf("hyperone: %w", err)
	}
	if len(records) == 1 {
		if records[0].Content != info.Value {
			return fmt.Errorf("hyperone: record with content %s not found: fqdn=%s", info.Value, info.EffectiveFQDN)
		}

		err = d.client.DeleteRecordset(zone.ID, recordset.ID)
		if err != nil {
			return fmt.Errorf("hyperone: failed to delete record: fqdn=%s, zone ID=%s, recordset ID=%s: %w", info.EffectiveFQDN, zone.ID, recordset.ID, err)
		}

		return nil
	}

	for _, record := range records {
		if record.Content == info.Value {
			err = d.client.DeleteRecord(zone.ID, recordset.ID, record.ID)
			if err != nil {
				return fmt.Errorf("hyperone: fqdn=%s, zone ID=%s, recordset ID=%s, record ID=%s: %w", info.EffectiveFQDN, zone.ID, recordset.ID, record.ID, err)
			}

			return nil
		}
	}

	return fmt.Errorf("hyperone: fqdn=%s, failed to find record with given value", info.EffectiveFQDN)
}

// getHostedZone gets the hosted zone.
func (d *DNSProvider) getHostedZone(fqdn string) (*internal.Zone, error) {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return nil, err
	}

	return d.client.FindZone(authZone)
}

func GetDefaultPassportLocation() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".h1", "passport.json"), nil
}
