package cloudflare

import (
	"context"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/libdns/cloudflare"
	"github.com/libdns/libdns"
)

// metaClient is a simple wrapper for the libdns/cloudflare Provider
type metaClient struct {
	client  *cloudflare.Provider
	zones   map[string]string // cache for zone name to zone ID mapping
	zonesMu *sync.RWMutex
}

// newClient creates a new metaClient instance
func newClient(config *Config) (*metaClient, error) {
	client := &cloudflare.Provider{
		AuthEmail: config.AuthEmail,
		AuthKey:   config.AuthKey,
		APIToken:  config.AuthToken,
		ZoneToken: config.ZoneToken,
		BaseURL:   config.BaseURL,
	}

	return &metaClient{
		client:  client,
		zones:   make(map[string]string),
		zonesMu: &sync.RWMutex{},
	}, nil
}

// CreateTXTRecord creates a new TXT record
func (m *metaClient) CreateTXTRecord(ctx context.Context, zoneName, recordName, value string, ttl int) (string, error) {
	// Create TXT record
	txt := libdns.TXT{
		Name: recordName,
		TTL:  time.Duration(ttl) * time.Second,
		Text: value,
	}

	// Convert TXT record to Record type
	record := libdns.Record(txt)

	records, err := m.client.SetRecords(ctx, zoneName, []libdns.Record{record})
	if err != nil {
		return "", err
	}

	if len(records) == 0 {
		return "", nil
	}

	// In this simplified version, we assume the record was created successfully, but don't care about the specific ID
	return "placeholder-id", nil
}

// DeleteTXTRecord deletes the specified TXT record
func (m *metaClient) DeleteTXTRecord(ctx context.Context, zoneName, recordName, value string) error {
	// Create TXT record
	txt := libdns.TXT{
		Name: recordName,
		Text: value,
	}

	// Convert TXT record to Record type
	record := libdns.Record(txt)

	_, err := m.client.DeleteRecords(ctx, zoneName, []libdns.Record{record})
	return err
}

// ZoneIDByName gets the zone ID
func (m *metaClient) ZoneIDByName(fdqn string) (string, error) {
	m.zonesMu.RLock()
	id := m.zones[fdqn]
	m.zonesMu.RUnlock()

	if id != "" {
		return id, nil
	}

	// libdns/cloudflare uses domain name instead of zone ID
	zone := dns01.UnFqdn(fdqn)

	m.zonesMu.Lock()
	m.zones[fdqn] = zone
	m.zonesMu.Unlock()

	return zone, nil
}
