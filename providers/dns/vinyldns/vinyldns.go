// Package vinyldns implements a DNS provider for solving the DNS-01 challenge using VinylDNS.
package vinyldns

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	"github.com/vinyldns/go-vinyldns/vinyldns"
)

// Environment variables names.
const (
	envNamespace = "VINYLDNS_"

	EnvAccessKey  = envNamespace + "ACCESS_KEY"
	EnvSecretKey  = envNamespace + "SECRET_KEY"
	EnvHost       = envNamespace + "HOST"
	EnvQuoteValue = envNamespace + "QUOTE_VALUE"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessKey  string
	SecretKey  string
	Host       string
	QuoteValue bool

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 30),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 4*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *vinyldns.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for VinylDNS.
// Credentials must be passed in the environment variables:
// VINYLDNS_ACCESS_KEY, VINYLDNS_SECRET_KEY, VINYLDNS_HOST.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessKey, EnvSecretKey, EnvHost)
	if err != nil {
		return nil, fmt.Errorf("vinyldns: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessKey = values[EnvAccessKey]
	config.SecretKey = values[EnvSecretKey]
	config.Host = values[EnvHost]
	config.QuoteValue = env.GetOrDefaultBool(EnvQuoteValue, false)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for VinylDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("vinyldns: the configuration of the VinylDNS DNS provider is nil")
	}

	if config.AccessKey == "" || config.SecretKey == "" {
		return nil, errors.New("vinyldns: credentials are missing")
	}

	if config.Host == "" {
		return nil, errors.New("vinyldns: host is missing")
	}

	client := vinyldns.NewClient(vinyldns.ClientConfiguration{
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		Host:      config.Host,
		UserAgent: useragent.Get(),
	})

	client.HTTPClient.Timeout = 30 * time.Second

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	existingRecord, err := d.getRecordSet(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("vinyldns: %w", err)
	}

	value := d.formatValue(info.Value)

	record := vinyldns.Record{Text: value}

	if existingRecord == nil || existingRecord.ID == "" {
		err = d.createRecordSet(ctx, info.EffectiveFQDN, []vinyldns.Record{record})
		if err != nil {
			return fmt.Errorf("vinyldns: %w", err)
		}

		return nil
	}

	for _, i := range existingRecord.Records {
		if i.Text == value {
			return nil
		}
	}

	records := existingRecord.Records
	records = append(records, record)

	err = d.updateRecordSet(ctx, existingRecord, records)
	if err != nil {
		return fmt.Errorf("vinyldns: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	existingRecord, err := d.getRecordSet(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("vinyldns: %w", err)
	}

	if existingRecord == nil || existingRecord.ID == "" || len(existingRecord.Records) == 0 {
		return nil
	}

	value := d.formatValue(info.Value)

	var records []vinyldns.Record
	for _, i := range existingRecord.Records {
		if i.Text != value {
			records = append(records, i)
		}
	}

	if len(records) == 0 {
		err = d.deleteRecordSet(ctx, existingRecord)
		if err != nil {
			return fmt.Errorf("vinyldns: %w", err)
		}

		return nil
	}

	err = d.updateRecordSet(ctx, existingRecord, records)
	if err != nil {
		return fmt.Errorf("vinyldns: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) formatValue(v string) string {
	if d.config.QuoteValue {
		return strconv.Quote(v)
	}

	return v
}
