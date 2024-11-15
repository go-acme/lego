// Package yandex360 implements a DNS provider for solving the DNS-01 challenge using Yandex 360.
package yandex360

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/yandex360/internal"
)

// Environment variables names.
const (
	envNamespace = "YANDEX360_"

	EnvOAuthToken = envNamespace + "OAUTH_TOKEN"
	EnvOrgID      = envNamespace + "ORG_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	OAuthToken         string
	OrgID              int64
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 21600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *internal.Client
	config *Config

	recordIDs   map[string]int64
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Yandex 360.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvOAuthToken, EnvOrgID)
	if err != nil {
		return nil, fmt.Errorf("yandex360: %w", err)
	}

	config := NewDefaultConfig()
	config.OAuthToken = values[EnvOAuthToken]

	orgID, err := strconv.ParseInt(values[EnvOrgID], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("yandex360: %w", err)
	}

	config.OrgID = orgID

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Yandex 360.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("yandex360: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.OAuthToken, config.OrgID)
	if err != nil {
		return nil, fmt.Errorf("yandex360: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		client:    client,
		config:    config,
		recordIDs: make(map[string]int64),
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("yandex360: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("yandex360: %w", err)
	}

	authZone = dns01.UnFqdn(authZone)

	record := internal.Record{
		Name: subDomain,
		TTL:  d.config.TTL,
		Text: info.Value,
		Type: "TXT",
	}

	newRecord, err := d.client.AddRecord(context.Background(), authZone, record)
	if err != nil {
		return fmt.Errorf("yandex360: add DNS record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("yandex360: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("yandex360: unknown recordID for %q", info.EffectiveFQDN)
	}

	err = d.client.DeleteRecord(context.Background(), authZone, recordID)
	if err != nil {
		return fmt.Errorf("yandex360: delete DNS record: %w", err)
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
