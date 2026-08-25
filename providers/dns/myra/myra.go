// Package myra implements a DNS provider for solving the DNS-01 challenge using Myra.
package myra

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Myra-Security-GmbH/myrasec-go/v2"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/internal/useragent"
	"github.com/go-acme/lego/v5/platform/env"
)

// Environment variables names.
const (
	envNamespace = "MYRA_"

	EnvAPIKey    = envNamespace + "API_KEY"
	EnvAPISecret = envNamespace + "API_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey    string
	APISecret string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 600*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 20*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *myrasec.API

	domainIDs   map[string]int
	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Myra.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("myra: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.APISecret = values[EnvAPISecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Myra.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("myra: the configuration of the DNS provider is nil")
	}

	client, err := myrasec.New(config.APIKey, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("myra: %w", err)
	}

	client.SetUserAgent(useragent.Get())

	return &DNSProvider{
		config:    config,
		client:    client,
		domainIDs: map[string]int{},
		recordIDs: map[string]int{},
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("myra: could not find zone for domain %q: %w", domain, err)
	}

	dom, err := d.client.FetchDomain(dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("myra: fetch domain %q: %w", domain, err)
	}

	if dom == nil {
		return fmt.Errorf("myra: no domain found for %q", authZone)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("myra: %w", err)
	}

	record := &myrasec.DNSRecord{
		Name:       subDomain,
		Value:      info.Value,
		RecordType: "TXT",
		Active:     true,
		Enabled:    true,
		TTL:        d.config.TTL,
	}

	newRecord, err := d.client.CreateDNSRecord(record, dom.ID)
	if err != nil {
		return fmt.Errorf("myra: create DNS record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.domainIDs[token] = dom.ID
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, recordOK := d.recordIDs[token]
	domainID, domainOK := d.domainIDs[token]
	d.recordIDsMu.Unlock()

	if !recordOK {
		return fmt.Errorf("myra: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	if !domainOK {
		return fmt.Errorf("myra: unknown domain ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	prevRecord, err := d.client.GetDNSRecord(domainID, recordID)
	if err != nil {
		return fmt.Errorf("myra: get previous record %d: %w", recordID, err)
	}

	rec := &myrasec.DNSRecord{
		ID:       prevRecord.ID,
		Modified: prevRecord.Modified,
	}

	_, err = d.client.DeleteDNSRecord(rec, domainID)
	if err != nil {
		return fmt.Errorf("myra: delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
