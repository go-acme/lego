// Package servercow implements a DNS provider for solving the DNS-01 challenge using Servercow DNS.
package servercow

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/servercow/internal"
)

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
		TTL:                env.GetOrDefaultInt(EnvTTL, 120),
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

	client := internal.NewClient(config.Username, config.Password)

	if config.HTTPClient == nil {
		client.HTTPClient = config.HTTPClient
	}

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
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := getAuthZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("servercow: %w", err)
	}

	ctx := context.Background()

	records, err := d.client.GetRecords(ctx, authZone)
	if err != nil {
		return fmt.Errorf("servercow: %w", err)
	}

	recordName, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("servercow: %w", err)
	}

	record := findRecords(records, recordName)

	// TXT record entry already existing
	if record != nil {
		if slices.Contains(record.Content, info.Value) {
			return nil
		}

		request := internal.Record{
			Name:    record.Name,
			TTL:     record.TTL,
			Type:    record.Type,
			Content: append(record.Content, info.Value),
		}

		_, err = d.client.CreateUpdateRecord(ctx, authZone, request)
		if err != nil {
			return fmt.Errorf("servercow: failed to update TXT records: %w", err)
		}
		return nil
	}

	request := internal.Record{
		Type:    "TXT",
		Name:    recordName,
		TTL:     d.config.TTL,
		Content: internal.Value{info.Value},
	}

	_, err = d.client.CreateUpdateRecord(ctx, authZone, request)
	if err != nil {
		return fmt.Errorf("servercow: failed to create TXT record %s: %w", info.EffectiveFQDN, err)
	}

	return nil
}

// CleanUp removes the TXT record previously created.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := getAuthZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("servercow: %w", err)
	}

	ctx := context.Background()

	records, err := d.client.GetRecords(ctx, authZone)
	if err != nil {
		return fmt.Errorf("servercow: failed to get TXT records: %w", err)
	}

	recordName, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("servercow: %w", err)
	}

	record := findRecords(records, recordName)
	if record == nil {
		return nil
	}

	if !slices.Contains(record.Content, info.Value) {
		return nil
	}

	// only 1 record value, the whole record must be deleted.
	if len(record.Content) == 1 {
		_, err = d.client.DeleteRecord(ctx, authZone, *record)
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
		if val != info.Value {
			request.Content = append(request.Content, val)
		}
	}

	_, err = d.client.CreateUpdateRecord(ctx, authZone, request)
	if err != nil {
		return fmt.Errorf("servercow: failed to update TXT records: %w", err)
	}

	return nil
}

func getAuthZone(domain string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return "", fmt.Errorf("could not find zone for FQDN %q: %w", domain, err)
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
