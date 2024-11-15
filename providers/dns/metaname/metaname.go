// Package metaname implements a DNS provider for solving the DNS-01 challenge using Metaname.
package metaname

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/nzdjb/go-metaname"
)

// Environment variables names.
const (
	envNamespace = "METANAME_"

	EnvAccountReference = envNamespace + "ACCOUNT_REFERENCE"
	EnvAPIKey           = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccountReference   string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *metaname.MetanameClient

	records   map[string]string
	recordsMu sync.Mutex
}

// NewDNSProvider returns a new DNS provider
// using environment variable METANAME_API_KEY for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccountReference, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("metaname: %w", err)
	}

	config := NewDefaultConfig()
	config.AccountReference = values[EnvAccountReference]
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Metaname.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("metaname: the configuration of the DNS provider is nil")
	}

	if config.AccountReference == "" {
		return nil, errors.New("metaname: missing account reference")
	}
	if config.APIKey == "" {
		return nil, errors.New("metaname: missing api key")
	}

	return &DNSProvider{
		config:  config,
		client:  metaname.NewMetanameClient(config.AccountReference, config.APIKey),
		records: make(map[string]string),
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("metaname: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("metaname: could not extract subDomain: %w", err)
	}

	ctx := context.Background()

	r := metaname.ResourceRecord{
		Name: subDomain,
		Type: "TXT",
		Aux:  nil,
		Ttl:  d.config.TTL,
		Data: info.Value,
	}

	ref, err := d.client.CreateDnsRecord(ctx, authZone, r)
	if err != nil {
		return fmt.Errorf("metaname: add record: %w", err)
	}

	d.recordsMu.Lock()
	d.records[token] = ref
	d.recordsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("metaname: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	ctx := context.Background()

	d.recordsMu.Lock()
	ref, ok := d.records[token]
	d.recordsMu.Unlock()

	if !ok {
		return fmt.Errorf("metaname: unknown ref for %s", info.EffectiveFQDN)
	}

	err = d.client.DeleteDnsRecord(ctx, authZone, ref)
	if err != nil {
		return fmt.Errorf("metaname: delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
