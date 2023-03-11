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

// DefaultEndpoint the default API endpoint.
const DefaultEndpoint = "https://api.zone.eu/v2/"

// Client the API client for Zoneee.
type Client struct {
	username string
	apiKey   string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(username string, apiKey string) *Client {
	baseURL, _ := url.Parse(DefaultEndpoint)

	return &Client{
		username:   username,
		apiKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetTxtRecords get TXT records.
// https://api.zone.eu/v2#operation/getdnstxtrecords
func (c *Client) GetTxtRecords(ctx context.Context, domain string) ([]TXTRecord, error) {
	endpoint := c.BaseURL.JoinPath("dns", domain, "txt")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}

	var records []TXTRecord
	if err := c.do(req, &records); err != nil {
		return nil, err
	}

	return records, nil
}

// AddTxtRecord creates a TXT records.
// https://api.zone.eu/v2#operation/creatednstxtrecord
func (c *Client) AddTxtRecord(ctx context.Context, domain string, record TXTRecord) ([]TXTRecord, error) {
	endpoint := c.BaseURL.JoinPath("dns", domain, "txt")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	var records []TXTRecord
	if err := c.do(req, &records); err != nil {
		return nil, err
	}

	return records, nil
}

// RemoveTxtRecord deletes a TXT record.
// https://api.zone.eu/v2#operation/deletednstxtrecord
func (c *Client) RemoveTxtRecord(ctx context.Context, domain, id string) error {
	endpoint := c.BaseURL.JoinPath("dns", domain, "txt", id)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	req.SetBasicAuth(c.username, c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
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
