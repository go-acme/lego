package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// defaultBaseURL represents the API endpoint to call.
const defaultBaseURL = "https://api.luadns.com"

// Client Lua DNS API client.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string

	apiUsername string
	apiToken    string
}

// NewClient creates a new Client.
func NewClient(apiUsername, apiToken string) *Client {
	return &Client{
		HTTPClient:  http.DefaultClient,
		BaseURL:     defaultBaseURL,
		apiUsername: apiUsername,
		apiToken:    apiToken,
	}
}

// ListZones gets all the hosted zones.
// https://luadns.com/api.html#list-zones
func (d *Client) ListZones() ([]DNSZone, error) {
	resp, err := d.do(http.MethodGet, "/v1/zones", nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		var errResp errorResponse
		err = json.Unmarshal(bodyBytes, &errResp)
		if err == nil {
			return nil, fmt.Errorf("api call error: Status=%v: %w", resp.StatusCode, errResp)
		}

		return nil, fmt.Errorf("api call error: Status=%d: %s", resp.StatusCode, string(bodyBytes))
	}

	var zones []DNSZone
	err = json.NewDecoder(resp.Body).Decode(&zones)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return zones, nil
}

// CreateRecord creates a new record in a zone.
// https://luadns.com/api.html#create-a-record
func (d *Client) CreateRecord(zone DNSZone, newRecord DNSRecord) (*DNSRecord, error) {
	body, err := json.Marshal(newRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resource := fmt.Sprintf("/v1/zones/%d/records", zone.ID)

	resp, err := d.do(http.MethodPost, resource, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		var errResp errorResponse
		err = json.Unmarshal(bodyBytes, &errResp)
		if err == nil {
			return nil, fmt.Errorf("could not create record %v: Status=%d: %w",
				string(body), resp.StatusCode, errResp)
		}

		return nil, fmt.Errorf("could not create record %v: Status=%d: %s",
			string(body), resp.StatusCode, string(bodyBytes))
	}

	var record *DNSRecord
	err = json.NewDecoder(resp.Body).Decode(&record)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return record, nil
}

// DeleteRecord deletes a record.
// https://luadns.com/api.html#delete-a-record
func (d *Client) DeleteRecord(record *DNSRecord) error {
	body, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	resource := fmt.Sprintf("/v1/zones/%d/records/%d", record.ZoneID, record.ID)

	resp, err := d.do(http.MethodDelete, resource, bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		var errResp errorResponse
		err = json.Unmarshal(bodyBytes, &errResp)
		if err == nil {
			return fmt.Errorf("could not delete record %v: Status=%d: %w",
				string(body), resp.StatusCode, errResp)
		}

		return fmt.Errorf("could not delete record %v: Status=%d: %s",
			string(body), resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (d *Client) do(method, uri string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", d.BaseURL, uri), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(d.apiUsername, d.apiToken)

	return d.HTTPClient.Do(req)
}
