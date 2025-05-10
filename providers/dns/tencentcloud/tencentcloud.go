// Package tencentcloud implements a DNS provider for solving the DNS-01 challenge using Tencent Cloud DNS.
package tencentcloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/libdns/libdns"
	"github.com/libdns/tencentcloud"
)

// Environment variables names.
const (
	envNamespace = "TENCENTCLOUD_"

	EnvSecretID     = envNamespace + "SECRET_ID"
	EnvSecretKey    = envNamespace + "SECRET_KEY"
	EnvRegion       = envNamespace + "REGION"
	EnvSessionToken = envNamespace + "SESSION_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	SecretID     string
	SecretKey    string
	Region       string
	SessionToken string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config   *Config
	provider *tencentcloud.Provider
}

// NewDNSProvider returns a DNSProvider instance configured for Tencent Cloud DNS.
// Credentials must be passed in the environment variable: TENCENTCLOUD_SECRET_ID, TENCENTCLOUD_SECRET_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvSecretID, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("tencentcloud: %w", err)
	}

	config := NewDefaultConfig()
	config.SecretID = values[EnvSecretID]
	config.SecretKey = values[EnvSecretKey]
	config.Region = env.GetOrDefaultString(EnvRegion, "")
	config.SessionToken = env.GetOrDefaultString(EnvSessionToken, "")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Tencent Cloud DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("tencentcloud: the configuration of the DNS provider is nil")
	}

	if config.SecretID == "" || config.SecretKey == "" {
		return nil, errors.New("tencentcloud: credentials missing")
	}

	provider := &tencentcloud.Provider{
		SecretId:     config.SecretID,
		SecretKey:    config.SecretKey,
		Region:       config.Region,
		SessionToken: config.SessionToken,
	}

	return &DNSProvider{config: config, provider: provider}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ttl := time.Duration(d.config.TTL) * time.Second

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to find zone: %w", err)
	}

	authZone = dns01.UnFqdn(authZone)

	recordName, err := extractRecordName(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to extract record name: %w", err)
	}

	record := libdns.RR{
		Type: "TXT",
		Name: recordName,
		Data: info.Value,
		TTL:  ttl,
	}

	recordVal, err := record.Parse()
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to parse TXT record: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.config.HTTPTimeout)
	defer cancel()

	_, err = d.provider.SetRecords(ctx, authZone, []libdns.Record{recordVal})
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to create TXT record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to find zone: %w", err)
	}

	authZone = dns01.UnFqdn(authZone)

	ctx, cancel := context.WithTimeout(context.Background(), d.config.HTTPTimeout)
	defer cancel()

	// Get all records for the zone
	records, err := d.provider.GetRecords(ctx, authZone)
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to get records: %w", err)
	}

	recordName, err := extractRecordName(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to extract record name: %w", err)
	}

	// Find the record we want to delete
	var recordsToDelete []libdns.Record
	for _, rec := range records {
		rr := rec.RR()
		if rr.Type == "TXT" && rr.Name == recordName && rr.Data == info.Value {
			recordsToDelete = append(recordsToDelete, rec)
		}
	}

	if len(recordsToDelete) == 0 {
		return nil
	}

	_, err = d.provider.DeleteRecords(ctx, authZone, recordsToDelete)
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to delete TXT record: %w", err)
	}

	return nil
}
