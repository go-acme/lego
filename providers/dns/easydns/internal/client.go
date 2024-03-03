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

// DefaultBaseURL the default API endpoint.
const DefaultBaseURL = "https://rest.easydns.net"

// Client the EasyDNS API client.
type Client struct {
	token string
	key   string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(token string, key string) *Client {
	baseURL, _ := url.Parse(DefaultBaseURL)

	return &Client{
		token:      token,
		key:        key,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) ListZones(ctx context.Context, domain string) ([]ZoneRecord, error) {
	endpoint := c.BaseURL.JoinPath("zones", "records", "all", domain)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	response := &apiResponse[[]ZoneRecord]{}
	err = c.do(req, response)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	return response.Data, nil
}

func (c *Client) AddRecord(ctx context.Context, domain string, record ZoneRecord) (string, error) {
	endpoint := c.BaseURL.JoinPath("zones", "records", "add", domain, "TXT")

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, record)
	if err != nil {
		return "", err
	}

	response := &apiResponse[*ZoneRecord]{}
	err = c.do(req, response)
	if err != nil {
		return "", err
	}

	if response.Error != nil {
		return "", response.Error
	}

	recordID := response.Data.ID

	return recordID, nil
}

func (c *Client) DeleteRecord(ctx context.Context, domain, recordID string) error {
	endpoint := c.BaseURL.JoinPath("zones", "records", domain, recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	err = c.do(req, nil)

	return err
}

func (c *Client) do(req *http.Request, result any) error {
	req.SetBasicAuth(c.token, c.key)

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

	query := endpoint.Query()
	query.Set("format", "json")
	endpoint.RawQuery = query.Encode()

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
