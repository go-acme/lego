package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// defaultBaseURL represents the API endpoint to call.
const defaultBaseURL = "https://api.luadns.com"

// Client Lua DNS API client.
type Client struct {
	apiUsername string
	apiToken    string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiUsername, apiToken string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiUsername: apiUsername,
		apiToken:    apiToken,
		baseURL:     baseURL,
		HTTPClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

// ListZones gets all the hosted zones.
// https://luadns.com/api.html#list-zones
func (c *Client) ListZones(ctx context.Context) ([]DNSZone, error) {
	endpoint := c.baseURL.JoinPath("v1", "zones")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var zones []DNSZone
	err = c.do(req, &zones)
	if err != nil {
		return nil, fmt.Errorf("could not list zones: %w", err)
	}

	return zones, nil
}

// CreateRecord creates a new record in a zone.
// https://luadns.com/api.html#create-a-record
func (c *Client) CreateRecord(ctx context.Context, zone DNSZone, newRecord DNSRecord) (*DNSRecord, error) {
	endpoint := c.baseURL.JoinPath("v1", "zones", strconv.Itoa(zone.ID), "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, newRecord)
	if err != nil {
		return nil, err
	}

	var record *DNSRecord
	err = c.do(req, &record)
	if err != nil {
		return nil, fmt.Errorf("could not create record %#v: %w", record, err)
	}

	return record, nil
}

// DeleteRecord deletes a record.
// https://luadns.com/api.html#delete-a-record
func (c *Client) DeleteRecord(ctx context.Context, record *DNSRecord) error {
	endpoint := c.baseURL.JoinPath("v1", "zones", strconv.Itoa(record.ZoneID), "records", strconv.Itoa(record.ID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, record)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return fmt.Errorf("could not delete record %#v: %w", record, err)
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.SetBasicAuth(c.apiUsername, c.apiToken)

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

	var errResp errorResponse
	err := json.Unmarshal(raw, &errResp)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("status=%d: %w", resp.StatusCode, errResp)
}
