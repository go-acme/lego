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

const defaultBaseURL = "https://api.nordname.fi/api/v3"

const customerUUIDHeader = "X-Customer-UUID"

// Client the Nordname API client.
type Client struct {
	apiKey       string
	customerUUID string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey, customerUUID string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:       apiKey,
		customerUUID: customerUUID,
		BaseURL:      baseURL,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// AddRecord adds a DNS record.
// https://app.nordname.eu/en/api-docs#tag/domain/POST/domain/{domain}/add_dns_record
func (c *Client) AddRecord(ctx context.Context, domain string, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("domain", domain, "add_dns_record")

	req, err := newRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := new(Record)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteRecord deletes a DNS record.
// https://app.nordname.eu/en/api-docs#tag/domain/DELETE/domain/{domain}/delete_dns_record
func (c *Client) DeleteRecord(ctx context.Context, domain, id string) error {
	endpoint := c.BaseURL.JoinPath("domain", domain, "delete_dns_record")

	query := endpoint.Query()
	query.Set("id", id)

	endpoint.RawQuery = query.Encode()

	req, err := newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set("Authorization", "token "+c.apiKey)

	if c.customerUUID != "" {
		req.Header.Set(customerUUIDHeader, c.customerUUID)
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

func newRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		values, err := querystring.Values(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request query parameters: %w", err)
		}

		endpoint.RawQuery = values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

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
