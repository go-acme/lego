// Package mittwald implements a DNS provider for solving the DNS-01 challenge using Mittwald.
package mittwald

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
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/mittwald/internal"
)

// Environment variables names.
const (
	envNamespace = "MITTWALD_"

	EnvToken = envNamespace + "TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const minTTL = 300

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token              string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, 2*time.Minute),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	zoneIDs   map[string]string
	zoneIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Mittwald.
// Credentials must be passed in the environment variables: MITTWALD_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken)
	if err != nil {
		return nil, fmt.Errorf("mittwald: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Mittwald.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("mittwald: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("mittwald: some credentials information are missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("mittwald: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := internal.NewClient(config.Token)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:  config,
		client:  client,
		zoneIDs: map[string]string{},
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getOrCreateZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("mittwald: get effective zone: %w", err)
	}

	record := internal.TXTRecord{
		Settings: internal.Settings{
			TTL: internal.TTL{Seconds: d.config.TTL},
		},
		Entries: []string{info.Value},
	}

	err = d.client.UpdateTXTRecord(ctx, zone.ID, record)
	if err != nil {
		return fmt.Errorf("mittwald: update/add TXT record: %w", err)
	}

	d.zoneIDsMu.Lock()
	d.zoneIDs[token] = zone.ID
	d.zoneIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// get the record's unique ID from when we created it
	d.zoneIDsMu.Lock()
	zoneID, ok := d.zoneIDs[token]
	d.zoneIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("mittwald: unknown zone ID for '%s'", info.EffectiveFQDN)
	}

	record := internal.TXTRecord{Entries: make([]string, 0)}

	err := d.client.UpdateTXTRecord(ctx, zoneID, record)
	if err != nil {
		return fmt.Errorf("mittwald: update/delete TXT record: %w", err)
	}

	return nil
}

func (d *DNSProvider) getOrCreateZone(ctx context.Context, fqdn string) (*internal.DNSZone, error) {
	domains, err := d.client.ListDomains(ctx)
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}

	dom, err := findDomain(domains, fqdn)
	if err != nil {
		return nil, fmt.Errorf("find domain: %w", err)
	}

	zones, err := d.client.ListDNSZones(ctx, dom.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("list DNS zones: %w", err)
	}

	for _, zone := range zones {
		if zone.Domain == dns01.UnFqdn(fqdn) {
			return &zone, nil
		}
	}

	// Looking for parent zone to create a new zone for the subdomain.

	parentZone, err := findZone(zones, fqdn)
	if err != nil {
		return nil, fmt.Errorf("find zone: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, parentZone.Domain)
	if err != nil {
		return nil, err
	}

	request := internal.CreateDNSZoneRequest{
		Name:         subDomain,
		ParentZoneID: parentZone.ID,
	}

	zone, err := d.client.CreateDNSZone(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("create DNS zone: %w", err)
	}

	return zone, nil
}

func findDomain(domains []internal.Domain, fqdn string) (internal.Domain, error) {
	for domain := range dns01.UnFqdnDomainsSeq(fqdn) {
		for _, dom := range domains {
			if dom.Domain == domain {
				return dom, nil
			}
		}
	}

	return internal.Domain{}, fmt.Errorf("domain %s not found", fqdn)
}

func findZone(zones []internal.DNSZone, fqdn string) (internal.DNSZone, error) {
	for domain := range dns01.UnFqdnDomainsSeq(fqdn) {
		for _, zon := range zones {
			if zon.Domain == domain {
				return zon, nil
			}
		}
	}

	return internal.DNSZone{}, fmt.Errorf("zone %s not found", fqdn)
}
