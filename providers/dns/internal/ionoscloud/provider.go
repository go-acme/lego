package ionoscloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	ionoscloud "github.com/go-acme/lego/v4/providers/dns/internal/ionoscloud/internal"
)

const MinTTL = 300

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *ionoscloud.Client
}

// NewDNSProviderConfig return a DNSProvider instance configured for IONOS Cloud.
func NewDNSProviderConfig(config *Config, baseURL string) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("credentials missing")
	}

	if config.TTL < MinTTL {
		return nil, fmt.Errorf("invalid TTL, TTL (%d) must be greater than %d", config.TTL, MinTTL)
	}

	client, err := ionoscloud.NewClient(config.Token)
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

	zoneName, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	zone, err := d.client.FindZone(ctx, dns01.UnFqdn(zoneName))
	if err != nil {
		return fmt.Errorf("failed to get zones: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zoneName)
	if err != nil {
		return fmt.Errorf("failed to extract subDomain from zone: %w", err)
	}

	record := ionoscloud.Record{
		Properties: ionoscloud.RecordProperties{
			Name:    subDomain,
			Content: info.Value,
			TTL:     d.config.TTL,
			Type:    "TXT",
			Enabled: true,
		},
	}

	err = d.client.CreateRecord(ctx, zone.ID, record)
	if err != nil {
		return fmt.Errorf("failed to create/update records (zone=%s): %w", zone.ID, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zoneName, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	zone, err := d.client.FindZone(ctx, dns01.UnFqdn(zoneName))
	if err != nil {
		return fmt.Errorf("failed to get zones: %w", err)
	}

	records, err := d.client.GetRecords(ctx, zone.ID)
	if err != nil {
		return fmt.Errorf("failed to get records (zone=%s): %w", zone.ID, err)
	}

	fqdn := dns01.UnFqdn(info.EffectiveFQDN)
	for _, record := range records {
		if record.MetaData.FQDN == fqdn && record.Properties.Content == strconv.Quote(info.Value) {
			err = d.client.RemoveRecord(ctx, zone.ID, record.ID)
			if err != nil {
				return fmt.Errorf("failed to remove record (zone=%s, record=%s): %w", zone.ID, record.ID, err)
			}

			return nil
		}
	}

	return fmt.Errorf("failed to remove record, record not found (zone=%s, domain=%s, fqdn=%s, value=%s)", zone.ID, domain, info.EffectiveFQDN, info.Value)
}
