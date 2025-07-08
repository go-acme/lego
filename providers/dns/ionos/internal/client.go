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
	querystring "github.com/google/go-querystring/query"
)

// defaultBaseURL represents the API endpoint to call.
const defaultBaseURL = "https://api.hosting.ionos.com/dns"

// APIKeyHeader API key header.
const APIKeyHeader = "X-Api-Key"

// Client Ionos API client.
type Client struct {
	apiKey string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey string) (*Client, error) {
	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		apiKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}, nil
}

// ListZones gets all zones.
func (c *Client) ListZones(ctx context.Context) ([]Zone, error) {
	endpoint := c.BaseURL.JoinPath("v1", "zones")

	req, err := makeJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var zones []Zone
	err = c.do(req, &zones)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}

	return zones, nil
}

// ReplaceRecords replaces some records of a zones.
func (c *Client) ReplaceRecords(ctx context.Context, zoneID string, records []Record) error {
	endpoint := c.BaseURL.JoinPath("v1", "zones", zoneID)

	req, err := makeJSONRequest(ctx, http.MethodPatch, endpoint, records)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	err = c.do(req, nil)
	if err != nil {
		return fmt.Errorf("failed to call API: %w", err)
	}

	return nil
}

// GetRecords gets the records of a zones.
func (c *Client) GetRecords(ctx context.Context, zoneID string, filter *RecordsFilter) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("v1", "zones", zoneID)

	req, err := makeJSONRequest(ctx, http.MethodGet, endpoint, nil)
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

	var zone CustomerZone
	err = c.do(req, &zone)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}

	return zone.Records, nil
}

// RemoveRecord removes a record.
func (c *Client) RemoveRecord(ctx context.Context, zoneID, recordID string) error {
	endpoint := c.BaseURL.JoinPath("v1", "zones", zoneID, "records", recordID)

	req, err := makeJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	err = c.do(req, nil)
	if err != nil {
		return fmt.Errorf("failed to call API: %w", err)
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set(APIKeyHeader, c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return parseError(req, resp)
	}

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func makeJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
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

	return req, nil
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	errClient := &ClientError{StatusCode: resp.StatusCode}
	err := json.Unmarshal(raw, &errClient.errors)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return errClient
}
