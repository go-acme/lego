package cloudflare

import (
	"context"
	"sync"

	"github.com/cloudflare/cloudflare-go"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

type metaClient struct {
	clientEdit *cloudflare.API // needs Zone/DNS/Edit permissions
	clientRead *cloudflare.API // needs Zone/Zone/Read permissions

	zones   map[string]string // caches calls to ZoneIDByName, see lookupZoneID()
	zonesMu *sync.RWMutex
}

func newClient(config *Config) (*metaClient, error) {
	// with AuthKey/AuthEmail we can access all available APIs
	if config.AuthToken == "" {
		client, err := cloudflare.New(config.AuthKey, config.AuthEmail, cloudflare.HTTPClient(config.HTTPClient))
		if err != nil {
			return nil, err
		}

		return &metaClient{
			clientEdit: client,
			clientRead: client,
			zones:      make(map[string]string),
			zonesMu:    &sync.RWMutex{},
		}, nil
	}

	dns, err := cloudflare.NewWithAPIToken(config.AuthToken, cloudflare.HTTPClient(config.HTTPClient))
	if err != nil {
		return nil, err
	}

	if config.ZoneToken == "" || config.ZoneToken == config.AuthToken {
		return &metaClient{
			clientEdit: dns,
			clientRead: dns,
			zones:      make(map[string]string),
			zonesMu:    &sync.RWMutex{},
		}, nil
	}

	zone, err := cloudflare.NewWithAPIToken(config.ZoneToken, cloudflare.HTTPClient(config.HTTPClient))
	if err != nil {
		return nil, err
	}

	return &metaClient{
		clientEdit: dns,
		clientRead: zone,
		zones:      make(map[string]string),
		zonesMu:    &sync.RWMutex{},
	}, nil
}

func (m *metaClient) CreateDNSRecord(ctx context.Context, zoneID string, rr cloudflare.DNSRecord) (*cloudflare.DNSRecordResponse, error) {
	return m.clientEdit.CreateDNSRecord(ctx, zoneID, rr)
}

func (m *metaClient) DNSRecords(ctx context.Context, zoneID string, rr cloudflare.DNSRecord) ([]cloudflare.DNSRecord, error) {
	return m.clientEdit.DNSRecords(ctx, zoneID, rr)
}

func (m *metaClient) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	return m.clientEdit.DeleteDNSRecord(ctx, zoneID, recordID)
}

func (m *metaClient) ZoneIDByName(fdqn string) (string, error) {
	m.zonesMu.RLock()
	id := m.zones[fdqn]
	m.zonesMu.RUnlock()

	if id != "" {
		return id, nil
	}

	id, err := m.clientRead.ZoneIDByName(dns01.UnFqdn(fdqn))
	if err != nil {
		return "", err
	}

	m.zonesMu.Lock()
	m.zones[fdqn] = id
	m.zonesMu.Unlock()
	return id, nil
}
