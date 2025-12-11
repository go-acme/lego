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
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

// defaultBaseURL is the default API endpoint.
const defaultBaseURL = "https://api.neodigit.net/v1"

// Client is a Tecnocrática API client.
type Client struct {
	token string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("credentials missing: token")
	}

	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		token:      token,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// GetZones lists all DNS zones.
func (c *Client) GetZones(ctx context.Context) ([]Zone, error) {
	endpoint := c.BaseURL.JoinPath("dns", "zones")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var zones []Zone

	err = c.do(req, &zones)
	if err != nil {
		return nil, err
	}

	return zones, nil
}

// GetRecords lists all records in a zone.
func (c *Client) GetRecords(ctx context.Context, zoneID int, recordType string) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("dns", "zones", strconv.Itoa(zoneID), "records")

	if recordType != "" {
		query := endpoint.Query()
		query.Set("type", recordType)
		endpoint.RawQuery = query.Encode()
	}

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var records []Record

	err = c.do(req, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// CreateRecord creates a new DNS record.
func (c *Client) CreateRecord(ctx context.Context, zoneID int, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("dns", "zones", strconv.Itoa(zoneID), "records")

	payload := RecordRequest{Record: record}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}

	var result Record

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteRecord deletes a DNS record.
func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordID int) error {
	endpoint := c.BaseURL.JoinPath("dns", "zones", strconv.Itoa(zoneID), "records", strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set("X-TCpanel-Token", c.token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
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
