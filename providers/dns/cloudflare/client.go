package cloudflare

import "github.com/cloudflare/cloudflare-go"

type metaClient struct {
	clientEdit *cloudflare.API // needs Zone/DNS/Edit permissions
	clientRead *cloudflare.API // needs Zone/Zone/Read permissions
}

func newClient(config *Config) (*metaClient, error) {
	// with AuthKey/AuthEmail we can access all available APIs
	if config.AuthToken == "" {
		client, err := cloudflare.New(config.AuthKey, config.AuthEmail, cloudflare.HTTPClient(config.HTTPClient))
		if err != nil {
			return nil, err
		}

		return &metaClient{clientEdit: client, clientRead: client}, nil
	}

	dns, err := cloudflare.NewWithAPIToken(config.AuthToken, cloudflare.HTTPClient(config.HTTPClient))
	if err != nil {
		return nil, err
	}

	if config.ZoneToken == "" || config.ZoneToken == config.AuthToken {
		return &metaClient{clientEdit: dns, clientRead: dns}, nil
	}

	zone, err := cloudflare.NewWithAPIToken(config.ZoneToken, cloudflare.HTTPClient(config.HTTPClient))
	if err != nil {
		return nil, err
	}

	return &metaClient{clientEdit: dns, clientRead: zone}, nil
}

func (m *metaClient) CreateDNSRecord(zoneID string, rr cloudflare.DNSRecord) (*cloudflare.DNSRecordResponse, error) {
	return m.clientEdit.CreateDNSRecord(zoneID, rr)
}

func (m *metaClient) DNSRecords(zoneID string, rr cloudflare.DNSRecord) ([]cloudflare.DNSRecord, error) {
	return m.clientEdit.DNSRecords(zoneID, rr)
}

func (m *metaClient) DeleteDNSRecord(zoneID, recordID string) error {
	return m.clientEdit.DeleteDNSRecord(zoneID, recordID)
}

func (m *metaClient) ZoneIDByName(zoneName string) (string, error) {
	return m.clientRead.ZoneIDByName(zoneName)
}
