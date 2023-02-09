package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// defaultBaseURL represents the API endpoint to call.
const defaultBaseURL = "https://dns.hetzner.com"

const authHeader = "Auth-API-Token"

// Client the Hetzner client.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string

	apiKey string
}

// NewClient Creates a new Hetzner client.
func NewClient(apiKey string) *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    defaultBaseURL,
		apiKey:     apiKey,
	}
}

// GetTxtRecord gets a TXT record.
func (c *Client) GetTxtRecord(name, value, zoneID string) (*DNSRecord, error) {
	records, err := c.getRecords(zoneID)
	if err != nil {
		return nil, err
	}

	for _, record := range records.Records {
		if record.Type == "TXT" && record.Name == name && record.Value == value {
			return &record, nil
		}
	}

	return nil, fmt.Errorf("could not find record: zone ID: %s; Record: %s", zoneID, name)
}

// https://dns.hetzner.com/api-docs#operation/GetRecords
func (c *Client) getRecords(zoneID string) (*DNSRecords, error) {
	endpoint, err := c.createEndpoint("api", "v1", "records")
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	query := endpoint.Query()
	query.Set("zone_id", zoneID)
	endpoint.RawQuery = query.Encode()

	resp, err := c.do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("could not get records: zone ID: %s; Status: %s; Body: %s",
			zoneID, resp.Status, string(bodyBytes))
	}

	records := &DNSRecords{}
	err = json.NewDecoder(resp.Body).Decode(records)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return records, nil
}

// CreateRecord creates a DNS record.
// https://dns.hetzner.com/api-docs#operation/CreateRecord
func (c *Client) CreateRecord(record DNSRecord) error {
	body, err := json.Marshal(record)
	if err != nil {
		return err
	}

	endpoint, err := c.createEndpoint("api", "v1", "records")
	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	resp, err := c.do(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("could not create record %s; Status: %s; Body: %s", string(body), resp.Status, string(bodyBytes))
	}

	return nil
}

// DeleteRecord deletes a DNS record.
// https://dns.hetzner.com/api-docs#operation/DeleteRecord
func (c *Client) DeleteRecord(recordID string) error {
	endpoint, err := c.createEndpoint("api", "v1", "records", recordID)
	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	resp, err := c.do(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not delete record: %s; Status: %s", resp.Status, recordID)
	}

	return nil
}

// GetZoneID gets the zone ID for a domain.
func (c *Client) GetZoneID(domain string) (string, error) {
	zones, err := c.getZones(domain)
	if err != nil {
		return "", err
	}

	for _, zone := range zones.Zones {
		if zone.Name == domain {
			return zone.ID, nil
		}
	}

	return "", fmt.Errorf("could not get zone for domain %s not found", domain)
}

// https://dns.hetzner.com/api-docs#operation/GetZones
func (c *Client) getZones(name string) (*Zones, error) {
	endpoint, err := c.createEndpoint("api", "v1", "zones")
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	query := endpoint.Query()
	query.Set("name", name)
	endpoint.RawQuery = query.Encode()

	resp, err := c.do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("could not get zones: %w", err)
	}

	// EOF fallback
	if resp.StatusCode == http.StatusNotFound {
		return &Zones{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not get zones: %s", resp.Status)
	}

	zones := &Zones{}
	err = json.NewDecoder(resp.Body).Decode(zones)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return zones, nil
}

func (c *Client) do(method string, endpoint fmt.Stringer, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(authHeader, c.apiKey)

	return c.HTTPClient.Do(req)
}

func (c *Client) createEndpoint(parts ...string) (*url.URL, error) {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	return baseURL.JoinPath(parts...), nil
}
