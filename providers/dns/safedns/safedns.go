// Package safedns implements a DNS provider for solving the DNS-01 challenge using UKFast SafeDNS.
package safedns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/safedns/internal"
)

// Environment variables.
const (
	envNamespace = "SAFEDNS_"

	EnvAuthToken = envNamespace + "AUTH_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AuthToken string

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
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

	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAuthToken)
	if err != nil {
		return nil, fmt.Errorf("safedns: %w", err)
	}

	config := NewDefaultConfig()
	config.AuthToken = values[EnvAuthToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for UKFast SafeDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("safedns: supplied configuration was nil")
	}

	if config.AuthToken == "" {
		return nil, errors.New("safedns: credentials missing")
	}

	client := internal.NewClient(config.AuthToken)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int),
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

	zone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("safedns: could not find zone for domain %q: %w", domain, err)
	}

	record := internal.Record{
		Name:    dns01.UnFqdn(info.EffectiveFQDN),
		Type:    "TXT",
		Content: fmt.Sprintf("%q", info.Value),
		TTL:     d.config.TTL,
	}

	resp, err := d.client.AddRecord(context.Background(), zone, record)
	if err != nil {
		return fmt.Errorf("safedns: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = resp.Data.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record previously created.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("safedns: could not find zone for domain %q: %w", domain, err)
	}

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("safedns: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	err = d.client.RemoveRecord(context.Background(), authZone, recordID)
	if err != nil {
		return fmt.Errorf("safedns: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}
