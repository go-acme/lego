// Package tecnocratica implements a DNS provider for solving the DNS-01 challenge using Tecnocrática.
package tecnocratica

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
	"github.com/go-acme/lego/v4/providers/dns/internal/tecnocratica/internal"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	zoneIDs     map[string]int
	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProviderConfig return a DNSProvider instance configured for Tecnocrática.
func NewDNSProviderConfig(config *Config, baseURL string) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("missing credentials")
	}

	client, err := internal.NewClient(config.Token)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	if baseURL != "" {
		client.BaseURL, err = url.Parse(baseURL)
		if err != nil {
			return nil, err
		}
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		zoneIDs:   make(map[string]int),
		recordIDs: make(map[string]int),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	zone, err := d.findZone(ctx, authZone)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	record := internal.Record{
		Name:    subDomain,
		Type:    "TXT",
		Content: info.Value,
		TTL:     d.config.TTL,
	}

	newRecord, err := d.client.CreateRecord(ctx, zone.ID, record)
	if err != nil {
		return fmt.Errorf("create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.zoneIDs[token] = zone.ID
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordIDsMu.Lock()
	zoneID, zoneOK := d.zoneIDs[token]
	recordID, recordOK := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !zoneOK || !recordOK {
		return fmt.Errorf("unknown record ID or zone ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecord(context.Background(), zoneID, recordID)
	if err != nil {
		return fmt.Errorf("delete record: fqdn=%s, zoneID=%d, recordID=%d: %w",
			info.EffectiveFQDN, zoneID, recordID, err)
	}

	d.recordIDsMu.Lock()
	delete(d.zoneIDs, token)
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) findZone(ctx context.Context, zoneName string) (*internal.Zone, error) {
	zones, err := d.client.GetZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("get zones: %w", err)
	}

	for _, zone := range zones {
		if zone.Name == zoneName || zone.HumanName == zoneName {
			return &zone, nil
		}
	}

	return nil, fmt.Errorf("zone not found: %s", zoneName)
}
