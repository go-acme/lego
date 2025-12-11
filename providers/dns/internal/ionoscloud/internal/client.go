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

const defaultBaseURL = "https://dns.de-fra.ionos.com"

// AuthorizationHeader bearer token header.
const AuthorizationHeader = "Authorization"

// Client IONOS Cloud Public DNS API client.
type Client struct {
	token string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(token string) (*Client, error) {
	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		token:      token,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}, nil
}

// FindZone get zone by zoneName.
func (c *Client) FindZone(ctx context.Context, zoneName string) (*Zone, error) {
	endpoint := c.BaseURL.JoinPath("zones")

	req, err := makeJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	query := req.URL.Query()
	query.Add("filter.zoneName", zoneName)
	req.URL.RawQuery = query.Encode()

	// IONOS Cloud API response wrapper
	var resp struct {
		Items []Zone `json:"items"`
	}

	err = c.do(req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}

	if len(resp.Items) != 1 {
		return nil, fmt.Errorf("no matching zone found for domain %q", zoneName)
	}

	return &resp.Items[0], nil
}

// CreateRecord Create a record in a zone.
func (c *Client) CreateRecord(ctx context.Context, zoneID string, record Record) error {
	endpoint := c.BaseURL.JoinPath("zones", zoneID, "records")

	req, err := makeJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	err = c.do(req, nil)
	if err != nil {
		return fmt.Errorf("failed to call API: %w", err)
	}

	return nil
}

// GetRecords gets the records of a zone.
func (c *Client) GetRecords(ctx context.Context, zoneID string) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("zones", zoneID, "records")

	req, err := makeJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Response wrapper for a single zone
	var resp struct {
		ID    string   `json:"id"`
		Type  string   `json:"type"`
		Items []Record `json:"items"`
	}

	err = c.do(req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}

	return resp.Items, nil
}

// RemoveRecord removes a record.
func (c *Client) RemoveRecord(ctx context.Context, zoneID, recordID string) error {
	endpoint := c.BaseURL.JoinPath("zones", zoneID, "records", recordID)

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
	if c.token != "" {
		req.Header.Set(AuthorizationHeader, "Bearer "+c.token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
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
