// Package huaweicloud implements a DNS provider for solving the DNS-01 challenge using Huawei Cloud.
package huaweicloud

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/platform/wait"
	"github.com/libdns/huaweicloud"
	"github.com/libdns/libdns"
)

// Environment variables names.
const (
	envNamespace = "HUAWEICLOUD_"

	EnvAccessKeyID     = envNamespace + "ACCESS_KEY_ID"
	EnvSecretAccessKey = envNamespace + "SECRET_ACCESS_KEY"
	EnvRegion          = envNamespace + "REGION"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int32
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                int32(env.GetOrDefaultInt(EnvTTL, 300)),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config   *Config
	provider *huaweicloud.Provider
}

// NewDNSProvider returns a DNSProvider instance configured for Huawei Cloud.
// Credentials must be passed in the environment variables:
// HUAWEICLOUD_ACCESS_KEY_ID, HUAWEICLOUD_SECRET_ACCESS_KEY
// HUAWEICLOUD_REGION is optional.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessKeyID, EnvSecretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("huaweicloud: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessKeyID = values[EnvAccessKeyID]
	config.SecretAccessKey = values[EnvSecretAccessKey]
	values, err = env.Get(EnvRegion)
	if err == nil {
		config.Region = values[EnvRegion]
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Huawei Cloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("huaweicloud: the configuration of the DNS provider is nil")
	}

	if config.AccessKeyID == "" || config.SecretAccessKey == "" {
		return nil, errors.New("huaweicloud: credentials missing")
	}

	provider := &huaweicloud.Provider{
		AccessKeyId:     config.AccessKeyID,
		SecretAccessKey: config.SecretAccessKey,
		RegionId:        config.Region,
	}

	return &DNSProvider{
		config:   config,
		provider: provider,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("huaweicloud: could not find zone for domain %q: %w", domain, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.config.HTTPTimeout)
	defer cancel()

	// Create TXT record
	record := libdns.TXT{
		Name: dns01.UnFqdn(info.EffectiveFQDN),
		TTL:  time.Duration(d.config.TTL) * time.Second,
		Text: info.Value,
	}

	_, err = d.provider.SetRecords(ctx, authZone, []libdns.Record{record})
	if err != nil {
		return fmt.Errorf("huaweicloud: %w", err)
	}

	// Wait for the record to propagate
	return wait.For("record propagation", d.config.PropagationTimeout, d.config.PollingInterval, func() (bool, error) {
		ctx, cancel := context.WithTimeout(context.Background(), d.config.HTTPTimeout)
		defer cancel()

		records, err := d.provider.GetRecords(ctx, authZone)
		if err != nil {
			return false, fmt.Errorf("get records: %w", err)
		}

		// Check if our record exists
		for _, record := range records {
			if txtRecord, ok := record.(libdns.TXT); ok {
				rr := txtRecord.RR()
				if rr.Type == "TXT" && rr.Name == dns01.UnFqdn(info.EffectiveFQDN) && txtRecord.Text == info.Value {
					return true, nil
				}
			}
		}

		return false, nil
	})
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("huaweicloud: could not find zone for domain %q: %w", domain, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.config.HTTPTimeout)
	defer cancel()

	// Find records matching our criteria
	records, err := d.provider.GetRecords(ctx, authZone)
	if err != nil {
		return fmt.Errorf("huaweicloud: %w", err)
	}

	var recordsToDelete []libdns.Record
	for _, record := range records {
		if txtRecord, ok := record.(libdns.TXT); ok {
			rr := txtRecord.RR()
			info.Value = fmt.Sprintf("\"%s\"", info.Value)
			if rr.Type == "TXT" && rr.Name == strings.TrimSuffix(dns01.UnFqdn(info.EffectiveFQDN), ".") && txtRecord.Text == info.Value {
				recordsToDelete = append(recordsToDelete, txtRecord)
			}
		}
	}

	if len(recordsToDelete) == 0 {
		return fmt.Errorf("huaweicloud: no record found to delete for %s", info.EffectiveFQDN)
	}

	_, err = d.provider.DeleteRecords(ctx, authZone, recordsToDelete)
	if err != nil {
		return fmt.Errorf("huaweicloud: failed to delete records: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
