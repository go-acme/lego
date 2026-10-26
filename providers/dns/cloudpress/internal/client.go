package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

const accountIDHeader = "X-Auth-Account"

// Client the CloudPress API client.
type Client struct {
	apiKey    string
	accountID string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(baseURL, apiKey, accountID string) (*Client, error) {
	if baseURL == "" {
		return nil, errors.New("missing base URL")
	}

	if apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL:  %w", err)
	}

	return &Client{
		apiKey:     apiKey,
		accountID:  accountID,
		baseURL:    base,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateRecord creates a DNS record.
// https://docs.cloudpress.com/api/dns/#create-record
func (c *Client) CreateRecord(ctx context.Context, zoneID string, record Record) (*Record, error) {
	endpoint := c.baseURL.JoinPath("api", "dns_zones", zoneID, "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, RecordEnvelope{Record: &record})
	if err != nil {
		return nil, err
	}

	result := new(RecordEnvelope)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Record, nil
}

// DeleteRecord deletes a DNS record.
// https://docs.cloudpress.com/api/dns/#delete-record
func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	endpoint := c.baseURL.JoinPath("api", "dns_zones", zoneID, "records", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// ListZones lists the DNS zones.
// https://docs.cloudpress.com/api/dns/#list-dns-zones
func (c *Client) ListZones(ctx context.Context, pager *Pager) ([]Zone, error) {
	endpoint := c.baseURL.JoinPath("api", "dns_zones")

	values, err := querystring.Values(pager)
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = values.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(ZonesResponse)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Zones, nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	if c.accountID != "" {
		req.Header.Set(accountIDHeader, c.accountID)
	}

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
