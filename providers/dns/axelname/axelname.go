// Package axelname implements a DNS provider for solving the DNS-01 challenge using Axelname.
package axelname

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/axelname/internal"
)

// Environment variables names.
const (
	envNamespace = "AXELNAME_"

	EnvNickname = envNamespace + "NICKNAME"
	EnvToken    = envNamespace + "TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Nickname string
	Token    string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
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

// NewDNSProvider returns a DNSProvider instance configured for Axelname.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvNickname, EnvToken)
	if err != nil {
		return nil, fmt.Errorf("axelname: %w", err)
	}

	config := NewDefaultConfig()
	config.Nickname = values[EnvNickname]
	config.Token = values[EnvToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Axelname.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("axelname: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.Nickname, config.Token)
	if err != nil {
		return nil, fmt.Errorf("axelname: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
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
		return fmt.Errorf("axelname: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("axelname: %w", err)
	}

	record := internal.Record{
		Name:  subDomain,
		Type:  "TXT",
		Value: info.Value,
	}

	err = d.client.AddRecord(context.Background(), dns01.UnFqdn(authZone), record)
	if err != nil {
		return fmt.Errorf("axelname: add record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("axelname: could not find zone for domain %q: %w", domain, err)
	}

	records, err := d.client.ListRecords(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("axelname: list records: %w", err)
	}

	for _, record := range records {
		if record.Type != "TXT" || record.Value != info.Value {
			continue
		}

		err = d.client.DeleteRecord(ctx, dns01.UnFqdn(authZone), record)
		if err != nil {
			return fmt.Errorf("axelname: delete record: %w", err)
		}

		return nil
	}

	return errors.New("axelname: delete record: record not found")
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
