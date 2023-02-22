// Package bunny implements a DNS provider for solving the DNS-01 challenge using Bunny DNS.
package bunny

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/simplesurance/bunny-go"
)

const minTTL = 60

// Environment variables names.
const (
	envNamespace = "BUNNY_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 2*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *bunny.Client
}

// NewDNSProvider returns a DNSProvider instance configured for bunny.
// Credentials must be passed in the environment variable: BUNNY_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("bunny: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for bunny.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("bunny: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("bunny: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("bunny: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := bunny.NewClient(config.APIKey)

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := getZone(fqdn)
	if err != nil {
		return fmt.Errorf("bunny: failed to find zone: fqdn=%s: %w", fqdn, err)
	}

	zones, err := d.client.DNSZone.List(context.Background(), &bunny.PaginationOptions{})
	if err != nil {
		return fmt.Errorf("bunny: %w", err)
	}

	var bunnyZone *bunny.DNSZone
	for _, z := range zones.Items {
		if *z.Domain == zone {
			bunnyZone = z
			break
		}
	}
	if bunnyZone == nil {
		return fmt.Errorf("bunny: could not find DNSZone zone=%s", zone)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, zone)
	if err != nil {
		return fmt.Errorf("bunny: %w", err)
	}

	typ := bunny.DNSRecordTypeTXT
	ttl := int32(d.config.TTL)
	opts := &bunny.AddOrUpdateDNSRecordOptions{
		Type:  &typ,
		Name:  &subDomain,
		Value: &value,
		TTL:   &ttl,
	}

	if _, err := d.client.DNSZone.AddDNSRecord(context.Background(), *bunnyZone.ID, opts); err != nil {
		return fmt.Errorf("bunny: failed to add TXT record: fqdn=%s, zoneID=%d: %w", fqdn, bunnyZone.ID, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zone, err := getZone(fqdn)
	if err != nil {
		return fmt.Errorf("bunny: failed to find zone: fqdn=%s: %w", fqdn, err)
	}

	zones, err := d.client.DNSZone.List(context.Background(), &bunny.PaginationOptions{})
	if err != nil {
		return fmt.Errorf("bunny: %w", err)
	}

	var bunnyZone *bunny.DNSZone
	for _, z := range zones.Items {
		if *z.Domain == zone {
			bunnyZone = z
			break
		}
	}
	if bunnyZone == nil {
		return fmt.Errorf("bunny: could not find DNSZone zone=%s", zone)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, zone)
	if err != nil {
		return fmt.Errorf("bunny: %w", err)
	}

	typ := bunny.DNSRecordTypeTXT
	var txtRecord *bunny.DNSRecord
	for _, record := range bunnyZone.Records {
		if *record.Name == subDomain && *record.Type == typ {
			record := record
			txtRecord = &record
			break
		}
	}
	if txtRecord == nil {
		return fmt.Errorf("bunny: could not find TXT record zone=%d, subdomain=%s", bunnyZone.ID, subDomain)
	}

	if err := d.client.DNSZone.DeleteDNSRecord(context.Background(), *bunnyZone.ID, *txtRecord.ID); err != nil {
		return fmt.Errorf("bunny: failed to delete TXT record: id=%d, name=%s: %w", txtRecord.ID, *txtRecord.Name, err)
	}

	return nil
}

func getZone(fqdn string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", err
	}

	return dns01.UnFqdn(authZone), nil
}
