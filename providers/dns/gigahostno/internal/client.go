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
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

const defaultBaseURL = "https://api.gigahost.no/api/v0"

const authorizationHeader = "Authorization"

// Client the Gigahost.no API client.
type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient() *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetZones returns all zones.
func (c *Client) GetZones(ctx context.Context) ([]Zone, error) {
	endpoint := c.BaseURL.JoinPath("dns", "zones")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result APIResponse[[]Zone]

	err = c.do(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetZoneRecords returns all records for a zone.
func (c *Client) GetZoneRecords(ctx context.Context, zoneID string) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("dns", "zones", zoneID, "records")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result APIResponse[[]Record]

	err = c.do(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// CreateNewRecord creates a new record.
func (c *Client) CreateNewRecord(ctx context.Context, zoneID string, record Record) error {
	endpoint := c.BaseURL.JoinPath("dns", "zones", zoneID, "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(ctx, req, nil)
}

// DeleteRecord deletes a record.
func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordID, name, recordType string) error {
	endpoint := c.BaseURL.JoinPath("dns", "zones", zoneID, "records", recordID)

	query := endpoint.Query()
	query.Set("name", name)
	query.Set("type", recordType)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(ctx, req, nil)
}

func (c *Client) do(ctx context.Context, req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set(authorizationHeader, "Bearer "+getToken(ctx))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
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

func newJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
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

	var errAPI APIError

	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return &errAPI
}
