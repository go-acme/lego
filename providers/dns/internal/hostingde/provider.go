// Package hostingde implements a DNS provider for solving the DNS-01 challenge using hosting.de.
package hostingde

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/internal/hostingde/internal"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	ZoneName           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProviderConfig return a DNSProvider instance configured for hosting.de.
func NewDNSProviderConfig(config *Config, baseURL string) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("API key missing")
	}

	client := internal.NewClient(config.APIKey)

	if baseURL != "" {
		client.BaseURL, _ = url.Parse(baseURL)
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
		return fmt.Errorf("could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	// get the ZoneConfig for that domain
	zonesFind := internal.ZoneConfigsFindRequest{
		Filter: internal.Filter{Field: "zoneName", Value: zoneName},
		Limit:  1,
		Page:   1,
	}

	zoneConfig, err := d.client.GetZone(ctx, zonesFind)
	if err != nil {
		return err
	}

	zoneConfig.Name = zoneName

	rec := []internal.DNSRecord{{
		Type:    "TXT",
		Name:    dns01.UnFqdn(info.EffectiveFQDN),
		Content: info.Value,
		TTL:     d.config.TTL,
	}}

	req := internal.ZoneUpdateRequest{
		ZoneConfig:   *zoneConfig,
		RecordsToAdd: rec,
	}

	response, err := d.client.UpdateZone(ctx, req)
	if err != nil {
		return err
	}

	for _, record := range response.Records {
		if record.Name == dns01.UnFqdn(info.EffectiveFQDN) && record.Content == fmt.Sprintf(`%q`, info.Value) {
			d.recordIDsMu.Lock()
			d.recordIDs[info.EffectiveFQDN] = record.ID
			d.recordIDsMu.Unlock()
		}
	}

	if d.recordIDs[info.EffectiveFQDN] == "" {
		return fmt.Errorf("error getting ID of just created record, for domain %s", domain)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneName, err := d.getZoneName(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	// get the ZoneConfig for that domain
	zonesFind := internal.ZoneConfigsFindRequest{
		Filter: internal.Filter{Field: "zoneName", Value: zoneName},
		Limit:  1,
		Page:   1,
	}

	zoneConfig, err := d.client.GetZone(ctx, zonesFind)
	if err != nil {
		return err
	}

	zoneConfig.Name = zoneName

	rec := []internal.DNSRecord{{
		Type:    "TXT",
		Name:    dns01.UnFqdn(info.EffectiveFQDN),
		Content: `"` + info.Value + `"`,
	}}

	req := internal.ZoneUpdateRequest{
		ZoneConfig:      *zoneConfig,
		RecordsToDelete: rec,
	}

	_, err = d.client.UpdateZone(ctx, req)
	if err != nil {
		return err
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, info.EffectiveFQDN)
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) getZoneName(fqdn string) (string, error) {
	if d.config.ZoneName != "" {
		return d.config.ZoneName, nil
	}

	zoneName, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", fmt.Errorf("could not find zone for %s: %w", fqdn, err)
	}

	if zoneName == "" {
		return "", errors.New("empty zone name")
	}

	return dns01.UnFqdn(zoneName), nil
}
