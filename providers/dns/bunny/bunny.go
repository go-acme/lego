// Package bunny implements a DNS provider for solving the DNS-01 challenge using Bunny DNS.
package bunny

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/miekg/dns"
	"github.com/nrdcg/bunny-go"
)

// Environment variables names.
const (
	envNamespace = "BUNNY_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

const minTTL = 60

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

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

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zone, err := d.findZone(ctx, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("bunny: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, deref(zone.Domain))
	if err != nil {
		return fmt.Errorf("bunny: %w", err)
	}

	record := &bunny.AddOrUpdateDNSRecordOptions{
		Type:  pointer(bunny.DNSRecordTypeTXT),
		Name:  pointer(subDomain),
		Value: pointer(info.Value),
		TTL:   pointer(int32(d.config.TTL)),
	}

	if _, err := d.client.DNSZone.AddDNSRecord(ctx, deref(zone.ID), record); err != nil {
		return fmt.Errorf("bunny: failed to add TXT record: fqdn=%s, zoneID=%d: %w", info.EffectiveFQDN, deref(zone.ID), err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zone, err := d.findZone(ctx, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("bunny: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, deref(zone.Domain))
	if err != nil {
		return fmt.Errorf("bunny: %w", err)
	}

	var record *bunny.DNSRecord
	for _, r := range zone.Records {
		if deref(r.Name) == subDomain && deref(r.Type) == bunny.DNSRecordTypeTXT {
			r := r
			record = &r
			break
		}
	}

	if record == nil {
		return fmt.Errorf("bunny: could not find TXT record zone=%d, subdomain=%s", deref(zone.ID), subDomain)
	}

	if err := d.client.DNSZone.DeleteDNSRecord(ctx, deref(zone.ID), deref(record.ID)); err != nil {
		return fmt.Errorf("bunny: failed to delete TXT record: id=%d, name=%s: %w", deref(record.ID), deref(record.Name), err)
	}

	return nil
}

func (d *DNSProvider) findZone(ctx context.Context, authZone string) (*bunny.DNSZone, error) {
	zones, err := d.client.DNSZone.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	domains := possibleDomains(authZone)

	var domainLength int

	var zone *bunny.DNSZone
	for _, item := range zones.Items {
		if item == nil {
			continue
		}

		curr := deref(item.Domain)

		if slices.Contains(domains, curr) && domainLength < len(curr) {
			domainLength = len(curr)

			zone = item
		}
	}

	if zone == nil {
		return nil, fmt.Errorf("could not find DNSZone zone=%s", authZone)
	}

	return zone, nil
}

func findZone(zones *bunny.DNSZones, authZone string) *bunny.DNSZone {
	domains := possibleDomains(authZone)

	var domainLength int

	var zone *bunny.DNSZone
	for _, item := range zones.Items {
		if item == nil {
			continue
		}

		curr := deref(item.Domain)

		if slices.Contains(domains, curr) && domainLength < len(curr) {
			domainLength = len(curr)

			zone = item
		}
	}

	return zone
}

func possibleDomains(fqdn string) []string {
	var d []string

	labelIndexes := dns.Split(fqdn)

	for i, index := range labelIndexes {
		if i == len(labelIndexes)-1 {
			continue
		}

		d = append(d, dns01.UnFqdn(fqdn[index:]))
	}

	return d
}

func pointer[T string | int | int32 | int64](v T) *T { return &v }

func deref[T string | int | int32 | int64](v *T) T {
	if v == nil {
		var zero T
		return zero
	}

	return *v
}
