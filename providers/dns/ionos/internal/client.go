package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	querystring "github.com/google/go-querystring/query"
)

// defaultBaseURL represents the API endpoint to call.
const defaultBaseURL = "https://api.hosting.ionos.com/dns"

// Client Ionos API client.
type Client struct {
	HTTPClient *http.Client
	BaseURL    *url.URL

	apiKey string
}

// NewClient creates a new Client.
func NewClient(apiKey string) (*Client, error) {
	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    baseURL,
		apiKey:     apiKey,
	}, nil
}

// ListZones gets all zones.
func (c *Client) ListZones(ctx context.Context) ([]Zone, error) {
	req, err := c.makeRequest(ctx, http.MethodGet, "/v1/zones", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp.Body, resp.StatusCode)
	}

	var zones []Zone
	err = json.NewDecoder(resp.Body).Decode(&zones)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return zones, nil
}

// ReplaceRecords replaces the some records of a zones.
func (c *Client) ReplaceRecords(ctx context.Context, zoneID string, records []Record) error {
	body, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := c.makeRequest(ctx, http.MethodPatch, path.Join("/v1/zones", zoneID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call API: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return readError(resp.Body, resp.StatusCode)
	}

	return nil
}

// GetRecords gets the records of a zones.
func (c *Client) GetRecords(ctx context.Context, zoneID string, filter *RecordsFilter) ([]Record, error) {
	req, err := c.makeRequest(ctx, http.MethodGet, path.Join("/v1/zones", zoneID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if filter != nil {
		v, errQ := querystring.Values(filter)
		if errQ != nil {
			return nil, errQ
		}

		req.URL.RawQuery = v.Encode()
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp.Body, resp.StatusCode)
	}

	var zone CustomerZone
	err = json.NewDecoder(resp.Body).Decode(&zone)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return zone.Records, nil
}

// RemoveRecord removes a record.
func (c *Client) RemoveRecord(ctx context.Context, zoneID, recordID string) error {
	req, err := c.makeRequest(ctx, http.MethodDelete, path.Join("/v1/zones", zoneID, "records", recordID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call API: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return readError(resp.Body, resp.StatusCode)
	}

	return nil
}

func (c *Client) makeRequest(ctx context.Context, method, uri string, body io.Reader) (*http.Request, error) {
	endpoint, err := c.BaseURL.Parse(path.Join(c.BaseURL.Path, uri))
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	return req, nil
}

func readError(body io.Reader, statusCode int) error {
	bodyBytes, _ := io.ReadAll(body)

	cErr := &ClientError{StatusCode: statusCode}

	err := json.Unmarshal(bodyBytes, &cErr.errors)
	if err != nil {
		cErr.message = string(bodyBytes)
		return cErr
	}

	return cErr
}
