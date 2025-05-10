// Package alidns implements a DNS provider for solving the DNS-01 challenge using Alibaba Cloud DNS.
package alidns

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/libdns/alidns"
	"github.com/libdns/libdns"
)

// Environment variables names.
const (
	envNamespace = "ALICLOUD_"

	EnvRAMRole       = envNamespace + "RAM_ROLE"
	EnvAccessKey     = envNamespace + "ACCESS_KEY"
	EnvSecretKey     = envNamespace + "SECRET_KEY"
	EnvSecurityToken = envNamespace + "SECURITY_TOKEN"
	EnvRegionID      = envNamespace + "REGION_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const defaultRegionID = "cn-hangzhou"

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	RAMRole            string
	APIKey             string
	SecretKey          string
	SecurityToken      string
	RegionID           string
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
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *alidns.Provider
}

// NewDNSProvider returns a DNSProvider instance configured for Alibaba Cloud DNS.
// - If you're using the instance RAM role, the RAM role environment variable must be passed in: ALICLOUD_RAM_ROLE.
// - Other than that, credentials must be passed in the environment variables:
// ALICLOUD_ACCESS_KEY, ALICLOUD_SECRET_KEY, and optionally ALICLOUD_SECURITY_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.RegionID = env.GetOrFile(EnvRegionID)

	values, err := env.Get(EnvRAMRole)
	if err == nil {
		config.RAMRole = values[EnvRAMRole]
	}

	values, err = env.Get(EnvAccessKey, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("alicloud: %w", err)
	}

	config.APIKey = values[EnvAccessKey]
	config.SecretKey = values[EnvSecretKey]
	config.SecurityToken = env.GetOrFile(EnvSecurityToken)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for alidns.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("alicloud: the configuration of the DNS provider is nil")
	}

	if config.RegionID == "" {
		config.RegionID = defaultRegionID
	}

	if config.APIKey == "" || config.SecretKey == "" {
		return nil, errors.New("alicloud: credentials missing")
	}

	client := &alidns.Provider{
		AccKeyID:      config.APIKey,
		AccKeySecret:  config.SecretKey,
		RegionID:      config.RegionID,
		RAMRole:       config.RAMRole,
		SecurityToken: config.SecurityToken,
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// Extract zone name from FQDN
	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("alicloud: %w", err)
	}

	zoneName := dns01.UnFqdn(authZone)
	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return err
	}

	txtRecord := libdns.TXT{
		Name: subDomain,
		Text: info.Value,
		TTL:  time.Duration(d.config.TTL) * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.config.HTTPTimeout)
	defer cancel()

	_, err = d.client.SetRecords(ctx, zoneName, []libdns.Record{txtRecord})
	if err != nil {
		return fmt.Errorf("alicloud: failed to create TXT record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// Extract zone name from FQDN
	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("alicloud: %w", err)
	}

	zoneName := dns01.UnFqdn(authZone)
	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.config.HTTPTimeout)
	defer cancel()

	// Get existing records to find the one we need to delete
	records, err := d.client.GetRecords(ctx, zoneName)
	if err != nil {
		return fmt.Errorf("alicloud: failed to get records: %w", err)
	}

	var recordsToDelete []libdns.Record
	for _, record := range records {
		txtRecord, ok := record.(libdns.TXT)
		if ok && txtRecord.Name == subDomain && txtRecord.Text == info.Value {
			recordsToDelete = append(recordsToDelete, txtRecord)
		}
	}

	if len(recordsToDelete) > 0 {
		_, err = d.client.DeleteRecords(ctx, zoneName, recordsToDelete)
		if err != nil {
			return fmt.Errorf("alicloud: failed to delete TXT record: %w", err)
		}
	}

	return nil
}
