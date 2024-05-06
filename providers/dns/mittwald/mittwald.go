// Package mittwald implements a DNS provider for solving the DNS-01 challenge using Mittwald.
package mittwald

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/mittwald/internal"
)

// Environment variables names.
const (
	envNamespace = "MITTWALD_"

	EnvToken = envNamespace + "TOKEN"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 2*time.Second),
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

	return &DNSProvider{
		config: config,
		client: internal.NewClient(config.Token),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	domains, err := d.client.ListDomains(ctx)
	if err != nil {
		return fmt.Errorf("mittwald: list domains: %w", err)
	}

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("mittwald: could not find zone for domain %q: %w", domain, err)
	}

	var projectID string
	for _, dom := range domains {
		// FIXME authZone or full domain? (ex: foo.bar.example.com)
		if dom.Domain == dns01.UnFqdn(authZone) {
			projectID = dom.ProjectID
			break
		}
	}

	if projectID == "" {
		return fmt.Errorf("mittwald: could not find project ID for domain %q (zone: %s)", domain, authZone)
	}

	zones, err := d.client.ListDNSZones(ctx, projectID)
	if err != nil {
		return fmt.Errorf("mittwald: list DNS zones: %w", err)
	}

	var parentZoneID string
	for _, zon := range zones {
		if zon.Domain == dns01.UnFqdn(authZone) {
			parentZoneID = zon.ID
			break
		}
	}

	if parentZoneID == "" {
		return fmt.Errorf("mittwald: could not find parent zone ID for domain %q (zone: %s)", domain, authZone)
	}

	// FIXME: if zone already exist? required to create multiple TXT records.

	// FIXME how to handle sub domain? (ex: foo.bar.example.com)
	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("mittwald: %w", err)
	}

	request := internal.CreateDNSZoneRequest{
		Name:         subDomain,
		ParentZoneID: parentZoneID,
	}

	zoneNew, err := d.client.CreateDNSZone(ctx, request)
	if err != nil {
		return fmt.Errorf("mittwald: create DNS zone: %w", err)
	}

	record := internal.TXTRecord{
		Settings: internal.Settings{
			TTL: internal.TTL{Auto: true},
		},
		Entries: []string{info.Value},
	}

	err = d.client.UpdateTXTRecord(ctx, zoneNew.ID, record)
	if err != nil {
		return fmt.Errorf("mittwald: update TXT record: %w", err)
	}

	d.zoneIDsMu.Lock()
	d.zoneIDs[token] = zoneNew.ID
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

	// FIXME update if zone has several TXT records.

	err := d.client.DeleteDNSZone(ctx, zoneID)
	if err != nil {
		return fmt.Errorf("mittwald: update delete DNS zone: %w", err)
	}

	return nil
}
