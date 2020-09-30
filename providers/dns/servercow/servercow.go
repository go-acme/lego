// Package servercow implements a DNS provider for solving the DNS-01 challenge using Servercow DNS.
package servercow

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/servercow/internal"
)

const defaultTTL = 120

// Environment variables names.
const (
	envNamespace = "SERVERCOW_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username string
	Password string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
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

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("servercow: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Servercow.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config.Username == "" || config.Password == "" {
		return nil, errors.New("servercow: incomplete credentials, missing username and/or password")
	}

	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}

	client := internal.NewClient(config.Username, config.Password)
	client.HTTPClient = config.HTTPClient

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.CreateRecord(domain, token, fqdn, value)
}

// CreateRecord creates a TXT record to fulfill the DNS-01 challenge.
func (d *DNSProvider) CreateRecord(domain, token, fqdn, value string) error {
	authZone, err := getAuthZone(domain)
	if err != nil {
		return fmt.Errorf("servercow: %w", err)
	}

	records, err := d.client.GetRecords(authZone)
	if err != nil {
		return fmt.Errorf("servercow: %w", err)
	}

	recordName := getRecordName(fqdn, authZone)

	record := findRecords(records, recordName)

	// TXT record entry already existing
	if record != nil {
		if containsValue(record, value) {
			return nil
		}

		request := internal.Record{
			Name:    record.Name,
			TTL:     record.TTL,
			Type:    record.Type,
			Content: append(record.Content, value),
		}

		_, err = d.client.CreateUpdateRecord(authZone, request)
		if err != nil {
			return fmt.Errorf("servercow: failed to update TXT records: %w", err)
		}
		return nil
	}

	request := internal.Record{
		Type:    "TXT",
		Name:    recordName,
		TTL:     d.config.TTL,
		Content: internal.Value{value},
	}

	_, err = d.client.CreateUpdateRecord(authZone, request)
	if err != nil {
		return fmt.Errorf("servercow: failed to create TXT record %s: %w", fqdn, err)
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
	authZone, err := getAuthZone(domain)
	if err != nil {
		return fmt.Errorf("servercow: %w", err)
	}

	records, err := d.client.GetRecords(authZone)
	if err != nil {
		return fmt.Errorf("servercow: failed to get TXT records: %w", err)
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
	if len(record.Content) == 1 {
		_, err = d.client.DeleteRecord(authZone, *record)
		if err != nil {
			return fmt.Errorf("servercow: failed to delete TXT records: %w", err)
		}
		return nil
	}

	request := internal.Record{
		Name: record.Name,
		Type: record.Type,
		TTL:  record.TTL,
	}

	for _, val := range record.Content {
		if val != value {
			request.Content = append(request.Content, val)
		}
	}

	_, err = d.client.CreateUpdateRecord(authZone, request)
	if err != nil {
		return fmt.Errorf("servercow: failed to update TXT records: %w", err)
	}

	return nil
}

func getAuthZone(domain string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return "", fmt.Errorf("could not find zone for domain %q: %w", domain, err)
	}

	zoneName := dns01.UnFqdn(authZone)
	return zoneName, nil
}

func findRecords(records []internal.Record, name string) *internal.Record {
	for _, r := range records {
		if r.Type == "TXT" && r.Name == name {
			return &r
		}
	}

	return nil
}

func containsValue(record *internal.Record, value string) bool {
	for _, val := range record.Content {
		if val == value {
			return true
		}
	}

	return false
}

func getRecordName(fqdn, authZone string) string {
	return fqdn[0 : len(fqdn)-len(authZone)-2]
}
