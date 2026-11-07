// Package nicru implements a DNS provider for solving the DNS-01 challenge using RU Center.
package nicru

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/nicru/internal"
	"golang.org/x/oauth2"
)

// Environment variables names.
const (
	envNamespace = "NICRU_"

	EnvUsername  = envNamespace + "USER"
	EnvPassword  = envNamespace + "PASSWORD"
	EnvServiceID = envNamespace + "SERVICE_ID"
	EnvSecret    = envNamespace + "SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	TTL                int
	Username           string
	Password           string
	ServiceID          string
	Secret             string
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
	config *Config

	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for RU Center.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword, EnvServiceID, EnvSecret)
	if err != nil {
		return nil, fmt.Errorf("nicru: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]
	config.ServiceID = values[EnvServiceID]
	config.Secret = values[EnvSecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for RU Center.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("nicru: the configuration of the DNS provider is nil")
	}

	err := validate(config)
	if err != nil {
		return nil, fmt.Errorf("nicru: %w", err)
	}

	authConfig := internal.CreateOAuthConfig("", config.ServiceID, config.Secret)

	return newDNSProviderConfig(config, authConfig)
}

func newDNSProviderConfig(config *Config, authConfig *oauth2.Config) (*DNSProvider, error) {
	ts := internal.NewTokenSource(context.Background(), authConfig, config.Username, config.Password)

	client, err := internal.NewClient(clientdebug.Wrap(oauth2.NewClient(context.Background(), ts)))
	if err != nil {
		return nil, fmt.Errorf("nicru: %w", err)
	}

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(ctx context.Context, domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("nicru: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	zone, err := findZone(ctx, d.client, authZone)
	if err != nil {
		return fmt.Errorf("nicru: find zone: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("nicru: %w", err)
	}

	records, err := d.client.GetRecords(ctx, zone.Service, authZone)
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
		TTL:  strconv.Itoa(d.config.TTL),
		Type: "TXT",
		TXT:  &internal.TXT{String: info.Value},
	}}

	_, err = d.client.AddRecords(ctx, zone.Service, authZone, rrs)
	if err != nil {
		return fmt.Errorf("nicru: add records: %w", err)
	}

	err = d.client.CommitZone(ctx, zone.Service, authZone)
	if err != nil {
		return fmt.Errorf("nicru: commit zone: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("nicru: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	zone, err := findZone(ctx, d.client, authZone)
	if err != nil {
		return fmt.Errorf("nicru: find zone: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("nicru: %w", err)
	}

	records, err := d.client.GetRecords(ctx, zone.Service, authZone)
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

		err = d.client.DeleteRecord(ctx, zone.Service, authZone, record.ID)
		if err != nil {
			return fmt.Errorf("nicru: delete record: %w", err)
		}
	}

	err = d.client.CommitZone(ctx, zone.Service, authZone)
	if err != nil {
		return fmt.Errorf("nicru: commit zone: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func findZone(ctx context.Context, client *internal.Client, authZone string) (*internal.Zone, error) {
	zones, err := client.ListZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch dns zones: %w", err)
	}

	if len(zones) == 0 {
		return nil, errors.New("no zones found")
	}

	for _, zone := range zones {
		if zone.Name == authZone {
			return &zone, nil
		}
	}

	return nil, fmt.Errorf("zone not found for %s", authZone)
}

func validate(config *Config) error {
	msg := " is missing in credentials information"

	if config.Username == "" {
		return errors.New("username" + msg)
	}

	if config.Password == "" {
		return errors.New("password" + msg)
	}

	if config.ServiceID == "" {
		return errors.New("serviceID" + msg)
	}

	if config.Secret == "" {
		return errors.New("secret" + msg)
	}

	return nil
}
