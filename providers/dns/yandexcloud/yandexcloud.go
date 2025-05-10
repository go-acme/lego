// Package yandexcloud implements a DNS provider for solving the DNS-01 challenge using Yandex Cloud.
package yandexcloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/yandexcloud/internal"
)

// Environment variables names.
const (
	envNamespace = "YANDEX_CLOUD_"

	EnvIamToken = envNamespace + "IAM_TOKEN"
	EnvFolderID = envNamespace + "FOLDER_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	IamToken string
	FolderID string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *internal.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Yandex Cloud.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvIamToken, EnvFolderID)
	if err != nil {
		return nil, fmt.Errorf("yandexcloud: %w", err)
	}

	config := NewDefaultConfig()
	config.IamToken = values[EnvIamToken]
	config.FolderID = values[EnvFolderID]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Yandex Cloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("yandexcloud: the configuration of the DNS provider is nil")
	}

	if config.IamToken == "" {
		return nil, errors.New("yandexcloud: some credentials information are missing IAM token")
	}

	if config.FolderID == "" {
		return nil, errors.New("yandexcloud: some credentials information are missing folder id")
	}

	client, err := internal.NewClient(config.IamToken, config.FolderID, config.TTL)
	if err != nil {
		return nil, fmt.Errorf("yandexcloud: %w", err)
	}

	return &DNSProvider{
		client: client,
		config: config,
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (r *DNSProvider) Present(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("yandexcloud: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	zones, err := r.client.ListZones(ctx)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	var zoneID string

	for _, zone := range zones {
		if zone.Zone == authZone {
			zoneID = zone.ID
		}
	}

	if zoneID == "" {
		return fmt.Errorf("yandexcloud: cant find dns zone %s in yandex cloud", authZone)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	err = r.client.CreateRecordSet(ctx, zoneID, subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (r *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("yandexcloud: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	zones, err := r.client.ListZones(ctx)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	var zoneID string

	for _, zone := range zones {
		if zone.Zone == authZone {
			zoneID = zone.ID
		}
	}

	if zoneID == "" {
		return nil
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	err = r.client.RemoveRecordSetValue(ctx, zoneID, subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (r *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return r.config.PropagationTimeout, r.config.PollingInterval
}
