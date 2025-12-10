// Package luadns implements a DNS provider for solving the DNS-01 challenge using LuaDNS.
package luadns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/luadns/internal"
)

// Environment variables names.
const (
	envNamespace = "LUADNS_"

	EnvAPIUsername = envNamespace + "API_USERNAME"
	EnvAPIToken    = envNamespace + "API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const minTTL = 300

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIUsername        string
	APIToken           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
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

	recordsMu sync.Mutex
	records   map[string]*internal.DNSRecord
}

// NewDNSProvider returns a DNSProvider instance configured for LuaDNS.
// Credentials must be passed in the environment variables:
// LUADNS_API_USERNAME and LUADNS_API_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIUsername, EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("luadns: %w", err)
	}

	config := NewDefaultConfig()
	config.APIUsername = values[EnvAPIUsername]
	config.APIToken = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for LuaDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("luadns: the configuration of the DNS provider is nil")
	}

	if config.APIUsername == "" || config.APIToken == "" {
		return nil, errors.New("luadns: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("luadns: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := internal.NewClient(config.APIUsername, config.APIToken)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:  config,
		client:  client,
		records: make(map[string]*internal.DNSRecord),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zones, err := d.client.ListZones(ctx)
	if err != nil {
		return fmt.Errorf("luadns: failed to get zones: %w", err)
	}

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("luadns: could not find zone for domain %q: %w", domain, err)
	}

	zone := findZone(zones, dns01.UnFqdn(authZone))
	if zone == nil {
		return fmt.Errorf("luadns: no matching zone found for domain %s", domain)
	}

	newRecord := internal.DNSRecord{
		Name:    info.EffectiveFQDN,
		Type:    "TXT",
		Content: info.Value,
		TTL:     d.config.TTL,
	}

	record, err := d.client.CreateRecord(ctx, *zone, newRecord)
	if err != nil {
		return fmt.Errorf("luadns: failed to create record: %w", err)
	}

	d.recordsMu.Lock()
	d.records[token] = record
	d.recordsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordsMu.Lock()
	record, ok := d.records[token]
	d.recordsMu.Unlock()

	if !ok {
		return fmt.Errorf("luadns: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	err := d.client.DeleteRecord(context.Background(), record)
	if err != nil {
		return fmt.Errorf("luadns: failed to delete record: %w", err)
	}

	// Delete record from map
	d.recordsMu.Lock()
	delete(d.records, token)
	d.recordsMu.Unlock()

	return nil
}

func findZone(zones []internal.DNSZone, domain string) *internal.DNSZone {
	var result *internal.DNSZone

	for _, zone := range zones {
		if zone.Name != "" && strings.HasSuffix(domain, zone.Name) {
			if result == nil || len(zone.Name) > len(result.Name) {
				result = &zone
			}
		}
	}

	return result
}
