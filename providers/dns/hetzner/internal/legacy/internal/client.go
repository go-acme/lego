package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// defaultBaseURL represents the API endpoint to call.
const defaultBaseURL = "https://dns.hetzner.com"

const authHeader = "Auth-API-Token"

// Client the Hetzner client.
type Client struct {
	apiKey string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Hetzner client.
func NewClient(apiKey string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetTxtRecord gets a TXT record.
func (c *Client) GetTxtRecord(ctx context.Context, name, value, zoneID string) (*DNSRecord, error) {
	records, err := c.getRecords(ctx, zoneID)
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
func (c *Client) getRecords(ctx context.Context, zoneID string) (*DNSRecords, error) {
	endpoint := c.baseURL.JoinPath("api", "v1", "records")

	query := endpoint.Query()
	query.Set("zone_id", zoneID)
	endpoint.RawQuery = query.Encode()

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	records := &DNSRecords{}
	err = json.Unmarshal(raw, records)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return records, nil
}

// CreateRecord creates a DNS record.
// https://dns.hetzner.com/api-docs#operation/CreateRecord
func (c *Client) CreateRecord(ctx context.Context, record DNSRecord) error {
	endpoint := c.baseURL.JoinPath("api", "v1", "records")

	req, err := c.newRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	return nil
}

// DeleteRecord deletes a DNS record.
// https://dns.hetzner.com/api-docs#operation/DeleteRecord
func (c *Client) DeleteRecord(ctx context.Context, recordID string) error {
	endpoint := c.baseURL.JoinPath("api", "v1", "records", recordID)

	req, err := c.newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	return nil
}

// GetZoneID gets the zone ID for a domain.
func (c *Client) GetZoneID(ctx context.Context, domain string) (string, error) {
	zones, err := c.getZones(ctx, domain)
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
func (c *Client) getZones(ctx context.Context, name string) (*Zones, error) {
	endpoint := c.baseURL.JoinPath("api", "v1", "zones")

	query := endpoint.Query()
	query.Set("name", name)
	endpoint.RawQuery = query.Encode()

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("could not get zones: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	// EOF fallback
	if resp.StatusCode == http.StatusNotFound {
		return &Zones{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	zones := &Zones{}
	err = json.Unmarshal(raw, zones)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return zones, nil
}

func (c *Client) newRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set(authHeader, c.apiKey)

	return req, nil
}
