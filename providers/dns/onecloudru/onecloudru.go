// Package onecloudru implements a DNS provider for solving the DNS-01 challenge using 1cloud.ru.
package onecloudru

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/onecloudru/internal"
)

// Environment variables names.
const (
	envNamespace = "ONECLOUDRU_"

	EnvToken = envNamespace + "TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
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

	recordIDs   map[string]int64
	recordIDsMu sync.Mutex

	domainIDs   map[string]int64
	domainIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for 1cloud.ru.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken)
	if err != nil {
		return nil, fmt.Errorf("onecloudru: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for 1cloud.ru.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("onecloudru: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.Token)
	if err != nil {
		return nil, fmt.Errorf("onecloudru: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int64),
		domainIDs: make(map[string]int64),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zone, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("onecloudru: %w", err)
	}

	d.domainIDsMu.Lock()
	d.domainIDs[token] = zone.ID
	d.domainIDsMu.Unlock()

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone.Name)
	if err != nil {
		return fmt.Errorf("onecloudru: %w", err)
	}

	ctrr := internal.CreateTXTRecordRequest{
		DomainID: strconv.FormatInt(zone.ID, 10),
		Name:     subDomain,
		TTL:      strconv.Itoa(internal.TTLRounder(d.config.TTL)),
		Text:     info.Value,
	}

	record, err := d.client.CreateTXTRecord(ctx, ctrr)
	if err != nil {
		return fmt.Errorf("onecloudru: create TXT record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = record.ID
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
		return fmt.Errorf("onecloudru: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	d.domainIDsMu.Lock()
	domainID, ok := d.domainIDs[token]
	d.domainIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("onecloudru: unknown domain ID for '%s'", info.EffectiveFQDN)
	}

	err := d.client.DeleteRecord(ctx, domainID, recordID)
	if err != nil {
		return fmt.Errorf("onecloudru: delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, fqdn string) (*internal.Domain, error) {
	domains, err := d.client.GetDomains(ctx)
	if err != nil {
		return nil, fmt.Errorf("get domains: %w", err)
	}

	for s := range dns01.UnFqdnDomainsSeq(fqdn) {
		for _, domain := range domains {
			if domain.Name == s {
				return &domain, nil
			}
		}
	}

	return nil, fmt.Errorf("no zone found for fqdn: %s", fqdn)
}
