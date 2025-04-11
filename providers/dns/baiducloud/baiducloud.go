// Package baiducloud implements a DNS provider for solving the DNS-01 challenge using Baidu Cloud.
package baiducloud

import (
	"errors"
	"fmt"
	"time"

	baidudns "github.com/baidubce/bce-sdk-go/services/dns"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/ptr"
)

// Environment variables names.
const (
	envNamespace = "BAIDUCLOUD_"

	EnvAccessKeyID     = envNamespace + "ACCESS_KEY_ID"
	EnvSecretAccessKey = envNamespace + "SECRET_ACCESS_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessKeyID     string
	SecretAccessKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *baidudns.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Baidu Cloud.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessKeyID, EnvSecretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("baiducloud: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessKeyID = values[EnvAccessKeyID]
	config.SecretAccessKey = values[EnvSecretAccessKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Baidu Cloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("baiducloud: the configuration of the DNS provider is nil")
	}

	if config.AccessKeyID == "" && config.SecretAccessKey == "" {
		return nil, errors.New("baiducloud: credentials missing")
	}

	client, err := baidudns.NewClient(config.AccessKeyID, config.SecretAccessKey, "")
	if err != nil {
		return nil, fmt.Errorf("baiducloud: %w", err)
	}

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("baiducloud: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("baiducloud: %w", err)
	}

	crr := &baidudns.CreateRecordRequest{
		Description: ptr.Pointer("lego"),
		Rr:          subDomain,
		Type:        "TXT",
		Value:       info.Value,
	}

	err = d.client.CreateRecord(dns01.UnFqdn(authZone), crr, "")
	if err != nil {
		return fmt.Errorf("baiducloud: create record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("baiducloud: could not find zone for domain %q: %w", domain, err)
	}

	lrr := &baidudns.ListRecordRequest{}

	recordResponse, err := d.client.ListRecord(dns01.UnFqdn(authZone), lrr)
	if err != nil {
		return fmt.Errorf("baiducloud: list record: %w", err)
	}

	recordID, err := findRecordID(recordResponse, info)
	if err != nil {
		return fmt.Errorf("baiducloud: find record: %w", err)
	}

	err = d.client.DeleteRecord(dns01.UnFqdn(authZone), recordID, "")
	if err != nil {
		return fmt.Errorf("baiducloud: delete record: %w", err)
	}

	return nil
}

func findRecordID(recordResponse *baidudns.ListRecordResponse, info dns01.ChallengeInfo) (string, error) {
	for _, record := range recordResponse.Records {
		if record.Type == "TXT" && record.Value == info.Value {
			return record.Id, nil
		}
	}

	return "", errors.New("record not found")
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
