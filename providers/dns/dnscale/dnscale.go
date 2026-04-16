// Package dnscale implements a DNS provider for solving the DNS-01 challenge using DNScale.
package dnscale

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/internal/env"
	"github.com/go-acme/lego/v5/providers/dns/dnscale/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "DNSCALE_"

	EnvAPIToken = envNamespace + "API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIToken string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
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

	zoneIDs     map[string]string
	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for DNScale.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("dnscale: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DNScale.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnscale: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIToken)
	if err != nil {
		return nil, fmt.Errorf("dnscale: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		zoneIDs:   make(map[string]string),
		recordIDs: make(map[string]string),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zoneID, err := d.findZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dnscale: %w", err)
	}

	record := internal.Record{
		Name:    dns01.UnFqdn(info.EffectiveFQDN),
		Type:    "TXT",
		Content: info.Value,
		TTL:     d.config.TTL,
	}

	newRecord, err := d.client.CreateRecord(ctx, zoneID, record)
	if err != nil {
		return fmt.Errorf("dnscale: %w", err)
	}

	d.recordIDsMu.Lock()
	d.zoneIDs[token] = zoneID
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	zoneID, zoneOK := d.zoneIDs[token]
	recordID, recordOK := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !zoneOK {
		return fmt.Errorf("bluecatv2: unknown zone ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	if !recordOK {
		return fmt.Errorf("bluecatv2: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecordByID(ctx, zoneID, recordID)
	if err != nil {
		return fmt.Errorf("dnscale: delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZoneID(ctx context.Context, fqdn string) (string, error) {
	var allZones []internal.Zone

	pager := &internal.Pager{Limit: 100}

	for {
		zonePage, err := d.client.ListZones(ctx, pager)
		if err != nil {
			return "", fmt.Errorf("list zones: %w", err)
		}

		allZones = append(allZones, zonePage.Items...)

		if len(zonePage.Items) < pager.Limit {
			break
		}

		pager.Offset += pager.Limit
	}

	for name := range dns01.DomainsSeq(fqdn) {
		for _, zone := range allZones {
			if strings.EqualFold(dns01.UnFqdn(name), dns01.UnFqdn(zone.Name)) {
				return zone.ID, nil
			}
		}
	}

	return "", fmt.Errorf("zone not found for domain %s", fqdn)
}
