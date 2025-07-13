package cloudflare

import (
	"context"
	"errors"
	"sync"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare/internal"
)

type metaClient struct {
	clientEdit *internal.Client // needs Zone/DNS/Edit permissions
	clientRead *internal.Client // needs Zone/Zone/Read permissions

	zones   map[string]string // caches calls to ZoneIDByName, see lookupZoneID()
	zonesMu *sync.RWMutex
}

func newClient(config *Config) (*metaClient, error) {
	// with AuthKey/AuthEmail we can access all available APIs
	if config.AuthToken == "" {
		client, err := internal.NewClient(
			internal.WithBaseURL(config.BaseURL),
			internal.WithHTTPClient(config.HTTPClient),
			internal.WithAuthKey(config.AuthEmail, config.AuthKey))
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

	dns, err := internal.NewClient(
		internal.WithBaseURL(config.BaseURL),
		internal.WithHTTPClient(config.HTTPClient),
		internal.WithAuthToken(config.AuthToken))
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

	zone, err := internal.NewClient(
		internal.WithBaseURL(config.BaseURL),
		internal.WithHTTPClient(config.HTTPClient),
		internal.WithAuthToken(config.ZoneToken))
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

func (m *metaClient) CreateDNSRecord(ctx context.Context, zoneID string, rr internal.Record) (*internal.Record, error) {
	return m.clientEdit.CreateDNSRecord(ctx, zoneID, rr)
}

func (m *metaClient) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	return m.clientEdit.DeleteDNSRecord(ctx, zoneID, recordID)
}

func (m *metaClient) ZoneIDByName(ctx context.Context, fdqn string) (string, error) {
	m.zonesMu.RLock()
	id := m.zones[fdqn]
	m.zonesMu.RUnlock()

	if id != "" {
		return id, nil
	}

	zones, err := m.clientRead.ZonesByName(ctx, dns01.UnFqdn(fdqn))
	if err != nil {
		return "", err
	}

	id, err = extractZoneID(zones)
	if err != nil {
		return "", err
	}

	m.zonesMu.Lock()
	m.zones[fdqn] = id
	m.zonesMu.Unlock()
	return id, nil
}

func extractZoneID(res []internal.Zone) (string, error) {
	switch len(res) {
	case 0:
		return "", errors.New("zone could not be found")
	case 1:
		return res[0].ID, nil
	default:
		return "", errors.New("ambiguous zone name; an account ID might help")
	}
}
