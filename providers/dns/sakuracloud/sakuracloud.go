// Package sakuracloud implements a DNS provider for solving the DNS-01 challenge using SakuraCloud DNS.
package sakuracloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/sakuracloud/internal"
)

// Environment variables names.
const (
	envNamespace = "SAKURACLOUD_"

	EnvAccessToken       = envNamespace + "ACCESS_TOKEN"
	EnvAccessTokenSecret = envNamespace + "ACCESS_TOKEN_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token              string
	Secret             string
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
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for SakuraCloud.
// Credentials must be passed in the environment variables:
// SAKURACLOUD_ACCESS_TOKEN & SAKURACLOUD_ACCESS_TOKEN_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessToken, EnvAccessTokenSecret)
	if err != nil {
		return nil, fmt.Errorf("sakuracloud: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvAccessToken]
	config.Secret = values[EnvAccessTokenSecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for SakuraCloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("sakuracloud: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("sakuracloud: AccessToken is missing")
	}

	if config.Secret == "" {
		return nil, errors.New("sakuracloud: AccessSecret is missing")
	}

	client, err := internal.NewClient(config.Token, config.Secret, config.TTL)
	if err != nil {
		return nil, fmt.Errorf("sakuracloud: %w", err)
	}

	return &DNSProvider{
		client: client,
		config: config,
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

	err := d.addTXTRecord(info.EffectiveFQDN, info.Value)
	if err != nil {
		return fmt.Errorf("sakuracloud: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	err := d.cleanupTXTRecord(info.EffectiveFQDN, info.Value)
	if err != nil {
		return fmt.Errorf("sakuracloud: %w", err)
	}

	return nil
}

// addTXTRecord adds a TXT record for the specified FQDN and value.
func (d *DNSProvider) addTXTRecord(fqdn string, value string) error {
	ctx := context.Background()

	// Get the appropriate zone for the domain
	domain := dns01.UnFqdn(fqdn)
	zone, err := d.client.GetZoneByDomain(ctx, domain)
	if err != nil {
		return fmt.Errorf("failed to find hosted zone: %w", err)
	}
	// Create the record
	recordName := getRecordName(fqdn, zone.Name)
	err = d.client.CreateTXTRecord(ctx, zone.ID, recordName, value)
	if err != nil {
		return fmt.Errorf("failed to create TXT record: %w", err)
	}

	return nil
}

// cleanupTXTRecord removes a TXT record for the specified FQDN and value.
func (d *DNSProvider) cleanupTXTRecord(fqdn string, value string) error {
	ctx := context.Background()

	// Get the appropriate zone for the domain
	domain := dns01.UnFqdn(fqdn)
	zone, err := d.client.GetZoneByDomain(ctx, domain)
	if err != nil {
		return fmt.Errorf("failed to find hosted zone: %w", err)
	}

	// Delete the record
	recordName := getRecordName(fqdn, zone.Name)
	err = d.client.DeleteTXTRecord(ctx, zone.ID, recordName, value)
	if err != nil {
		return fmt.Errorf("failed to delete TXT record: %w", err)
	}

	return nil
}

// getRecordName returns the name of the TXT record to create, relative to the zone.
func getRecordName(fqdn, domain string) string {
	name := dns01.UnFqdn(fqdn)

	if name == domain {
		return "@"
	}

	return name[:len(name)-len(domain)-1]
}
