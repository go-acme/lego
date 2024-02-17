// Package nicru implements a DNS provider for solving the DNS-01 challenge using RU Center.
package nicru

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/nicru/internal"
)

// Environment variables names.
const (
	envNamespace = "NICRU_"

	EnvUsername    = envNamespace + "USER"
	EnvPassword    = envNamespace + "PASSWORD"
	EnvServiceID   = envNamespace + "SERVICE_ID"
	EnvSecret      = envNamespace + "SECRET"
	EnvServiceName = envNamespace + "SERVICE_NAME"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	TTL                int
	Username           string
	Password           string
	ServiceID          string
	Secret             string
	ServiceName        string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 30),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 10*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 1*time.Minute),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *internal.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for RU Center.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword, EnvServiceID, EnvSecret, EnvServiceName)
	if err != nil {
		return nil, fmt.Errorf("nicru: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]
	config.ServiceID = values[EnvServiceID]
	config.Secret = values[EnvSecret]
	config.ServiceName = values[EnvServiceName]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for RU Center.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("nicru: the configuration of the DNS provider is nil")
	}

	clientCfg := &internal.OauthConfiguration{
		OAuth2ClientID: config.ServiceID,
		OAuth2SecretID: config.Secret,
		Username:       config.Username,
		Password:       config.Password,
	}

	oauthClient, err := internal.NewOauthClient(context.Background(), clientCfg)
	if err != nil {
		return nil, fmt.Errorf("nicru: %w", err)
	}

	client, err := internal.NewClient(oauthClient, config.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("nicru: unable to build API client: %w", err)
	}

	return &DNSProvider{
		client: client,
		config: config,
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (p *DNSProvider) Present(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("nicru: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	ctx := context.Background()

	err = p.checkZoneUUID(ctx, authZone)
	if err != nil {
		return fmt.Errorf("nicru: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("nicru: %w", err)
	}

	records, err := p.client.GetRecords(ctx, authZone)
	if err != nil {
		return fmt.Errorf("nicru: get records: %w", err)
	}

	for _, record := range records {
		if record.TXT == nil {
			continue
		}

		if record.TXT.Text == subDomain && record.TXT.String == info.Value {
			return nil
		}
	}

	rrs := []internal.RR{{
		Name: subDomain,
		TTL:  strconv.Itoa(p.config.TTL),
		Type: "TXT",
		TXT:  &internal.TXT{String: info.Value},
	}}

	_, err = p.client.AddRecords(ctx, authZone, rrs)
	if err != nil {
		return fmt.Errorf("nicru: add records: %w", err)
	}

	err = p.client.CommitZone(ctx, authZone)
	if err != nil {
		return fmt.Errorf("nicru: commit zone: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (p *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("nicru: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	ctx := context.Background()

	err = p.checkZoneUUID(ctx, authZone)
	if err != nil {
		return fmt.Errorf("nicru: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("nicru: %w", err)
	}

	records, err := p.client.GetRecords(ctx, authZone)
	if err != nil {
		return fmt.Errorf("nicru: get records: %w", err)
	}

	subDomain = dns01.UnFqdn(subDomain)

	for _, record := range records {
		if record.TXT == nil {
			continue
		}

		if record.Name != subDomain || record.TXT.String != info.Value {
			continue
		}

		err = p.client.DeleteRecord(ctx, authZone, record.ID)
		if err != nil {
			return fmt.Errorf("nicru: delete record: %w", err)
		}
	}

	err = p.client.CommitZone(ctx, authZone)
	if err != nil {
		return fmt.Errorf("nicru: commit zone: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (p *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return p.config.PropagationTimeout, p.config.PollingInterval
}

func (p *DNSProvider) checkZoneUUID(ctx context.Context, authZone string) error {
	zones, err := p.client.GetZones(ctx)
	if err != nil {
		return fmt.Errorf("unable to fetch dns zones: %w", err)
	}

	var zoneUUID string
	for _, zone := range zones {
		if zone.Name == authZone {
			zoneUUID = zone.ID
		}
	}

	if zoneUUID == "" {
		return fmt.Errorf("zone UUID not found for %s", authZone)
	}

	return nil
}
