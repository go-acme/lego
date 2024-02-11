// Package hostingde implements a DNS provider for solving the DNS-01 challenge using hosting.de.
package hostingde

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/hostingde"
)

// Environment variables names.
const (
	envNamespace = "HOSTINGDE_"

	EnvAPIKey   = envNamespace + "API_KEY"
	EnvZoneName = envNamespace + "ZONE_NAME"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	ZoneName           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
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
	client *hostingde.Client

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for hosting.de.
// Credentials must be passed in the environment variables:
// HOSTINGDE_ZONE_NAME and HOSTINGDE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("hostingde: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.ZoneName = env.GetOrFile(EnvZoneName)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for hosting.de.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("hostingde: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("hostingde: API key missing")
	}

	return &DNSProvider{
		config:    config,
		client:    hostingde.NewClient(config.APIKey),
		recordIDs: make(map[string]string),
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

	zoneName, err := d.getZoneName(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hostingde: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	// get the ZoneConfig for that domain
	zonesFind := hostingde.ZoneConfigsFindRequest{
		Filter: hostingde.Filter{Field: "zoneName", Value: zoneName},
		Limit:  1,
		Page:   1,
	}

	zoneConfig, err := d.client.GetZone(ctx, zonesFind)
	if err != nil {
		return fmt.Errorf("hostingde: %w", err)
	}

	zoneConfig.Name = zoneName

	rec := []hostingde.DNSRecord{{
		Type:    "TXT",
		Name:    dns01.UnFqdn(info.EffectiveFQDN),
		Content: info.Value,
		TTL:     d.config.TTL,
	}}

	req := hostingde.ZoneUpdateRequest{
		ZoneConfig:   *zoneConfig,
		RecordsToAdd: rec,
	}

	response, err := d.client.UpdateZone(ctx, req)
	if err != nil {
		return fmt.Errorf("hostingde: %w", err)
	}

	for _, record := range response.Records {
		if record.Name == dns01.UnFqdn(info.EffectiveFQDN) && record.Content == fmt.Sprintf(`%q`, info.Value) {
			d.recordIDsMu.Lock()
			d.recordIDs[info.EffectiveFQDN] = record.ID
			d.recordIDsMu.Unlock()
		}
	}

	if d.recordIDs[info.EffectiveFQDN] == "" {
		return fmt.Errorf("hostingde: error getting ID of just created record, for domain %s", domain)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneName, err := d.getZoneName(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hostingde: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	// get the ZoneConfig for that domain
	zonesFind := hostingde.ZoneConfigsFindRequest{
		Filter: hostingde.Filter{Field: "zoneName", Value: zoneName},
		Limit:  1,
		Page:   1,
	}

	zoneConfig, err := d.client.GetZone(ctx, zonesFind)
	if err != nil {
		return fmt.Errorf("hostingde: %w", err)
	}
	zoneConfig.Name = zoneName

	rec := []hostingde.DNSRecord{{
		Type:    "TXT",
		Name:    dns01.UnFqdn(info.EffectiveFQDN),
		Content: `"` + info.Value + `"`,
	}}

	req := hostingde.ZoneUpdateRequest{
		ZoneConfig:      *zoneConfig,
		RecordsToDelete: rec,
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, info.EffectiveFQDN)
	d.recordIDsMu.Unlock()

	_, err = d.client.UpdateZone(ctx, req)
	if err != nil {
		return fmt.Errorf("hostingde: %w", err)
	}
	return nil
}

func (d *DNSProvider) getZoneName(fqdn string) (string, error) {
	if d.config.ZoneName != "" {
		return d.config.ZoneName, nil
	}

	zoneName, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", fmt.Errorf("could not find zone: %w", err)
	}

	if zoneName == "" {
		return "", errors.New("empty zone name")
	}

	return dns01.UnFqdn(zoneName), nil
}
