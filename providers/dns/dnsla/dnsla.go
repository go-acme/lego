// Package dnsla implements a DNS provider for solving the DNS-01 challenge using dns.la.
package dnsla

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/dnsla/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "DNSLA_"

	EnvAPIID     = envNamespace + "API_ID"
	EnvAPISecret = envNamespace + "API_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIID     string
	APISecret string

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

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for dns.la.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIID, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("dnsla: %w", err)
	}

	config := NewDefaultConfig()
	config.APIID = values[EnvAPIID]
	config.APISecret = values[EnvAPISecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for dns.la.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnsla: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIID, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("dnsla: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]string),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dnsla: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("dnsla: %w", err)
	}

	zone, err := d.findDomain(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dnsla: %w", err)
	}

	record := internal.Record{
		DomainID: zone.ID,
		Type:     internal.TypeTXT,
		Host:     subDomain,
		Data:     info.Value,
		TTL:      d.config.TTL,
	}

	recordID, err := d.client.AddRecord(ctx, record)
	if err != nil {
		return fmt.Errorf("dnsla: add record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("dnsla: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	err := d.client.DeleteRecord(ctx, recordID)
	if err != nil {
		return fmt.Errorf("dnsla: delete record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findDomain(ctx context.Context, fqdn string) (internal.Domain, error) {
	pager := internal.Pager{
		PageIndex: 1,
		PageSize:  100,
	}

	domains, err := d.client.ListDomains(ctx, pager)
	if err != nil {
		return internal.Domain{}, fmt.Errorf("list domains: %w", err)
	}

	for d := range dns01.DomainsSeq(fqdn) {
		for _, domain := range domains {
			if domain.Domain == d {
				return domain, nil
			}
		}
	}

	return internal.Domain{}, fmt.Errorf("domain not found (fqdn: %q)", fqdn)
}
