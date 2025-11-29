package ionos

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	ionos "github.com/go-acme/lego/v4/providers/dns/internal/ionos/internal"
)

const MinTTL = 300

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *ionos.Client
}

// NewDNSProviderConfig return a DNSProvider instance configured for Ionos.
func NewDNSProviderConfig(config *Config, baseURL string) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("credentials missing")
	}

	if config.TTL < MinTTL {
		return nil, fmt.Errorf("invalid TTL, TTL (%d) must be greater than %d", config.TTL, MinTTL)
	}

	client, err := ionos.NewClient(config.APIKey)
	if err != nil {
		return nil, err
	}

	if baseURL != "" {
		client.BaseURL, _ = url.Parse(baseURL)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zones, err := d.client.ListZones(ctx)
	if err != nil {
		return fmt.Errorf("failed to get zones: %w", err)
	}

	name := dns01.UnFqdn(info.EffectiveFQDN)

	zone := findZone(zones, name)
	if zone == nil {
		return errors.New("no matching zone found for domain")
	}

	filter := &ionos.RecordsFilter{
		Suffix:     name,
		RecordType: "TXT",
	}

	records, err := d.client.GetRecords(ctx, zone.ID, filter)
	if err != nil {
		return fmt.Errorf("failed to get records (zone=%s): %w", zone.ID, err)
	}

	records = append(records, ionos.Record{
		Name:    name,
		Content: info.Value,
		TTL:     d.config.TTL,
		Type:    "TXT",
	})

	err = d.client.ReplaceRecords(ctx, zone.ID, records)
	if err != nil {
		return fmt.Errorf("failed to create/update records (zone=%s): %w", zone.ID, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zones, err := d.client.ListZones(ctx)
	if err != nil {
		return fmt.Errorf("failed to get zones: %w", err)
	}

	name := dns01.UnFqdn(info.EffectiveFQDN)

	zone := findZone(zones, name)
	if zone == nil {
		return errors.New("no matching zone found for domain")
	}

	filter := &ionos.RecordsFilter{
		Suffix:     name,
		RecordType: "TXT",
	}

	records, err := d.client.GetRecords(ctx, zone.ID, filter)
	if err != nil {
		return fmt.Errorf("failed to get records (zone=%s): %w", zone.ID, err)
	}

	for _, record := range records {
		if record.Name == name && record.Content == strconv.Quote(info.Value) {
			err = d.client.RemoveRecord(ctx, zone.ID, record.ID)
			if err != nil {
				return fmt.Errorf("failed to remove record (zone=%s, record=%s): %w", zone.ID, record.ID, err)
			}

			return nil
		}
	}

	return fmt.Errorf("failed to remove record, record not found (zone=%s, domain=%s, fqdn=%s, value=%s)", zone.ID, domain, info.EffectiveFQDN, info.Value)
}

func findZone(zones []ionos.Zone, domain string) *ionos.Zone {
	var result *ionos.Zone

	for _, zone := range zones {
		if zone.Name != "" && strings.HasSuffix(domain, zone.Name) {
			if result == nil || len(zone.Name) > len(result.Name) {
				result = &zone
			}
		}
	}

	return result
}
