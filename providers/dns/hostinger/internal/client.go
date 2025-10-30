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

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://developers.hostinger.com"

const authorizationHeader = "Authorization"

// Client the Hostinger API client.
type Client struct {
	token string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		token:      token,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// GetDNSRecords retrieves DNS zone records for a specific domain.
// https://developers.hostinger.com/#tag/dns-zone/get/api/dns/v1/zones/{domain}
func (c *Client) GetDNSRecords(ctx context.Context, domain string) ([]RecordSet, error) {
	endpoint := c.BaseURL.JoinPath("/api/dns/v1/zones/", domain)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result []RecordSet

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateDNSRecords updates DNS records for the selected domain.
// https://developers.hostinger.com/#tag/dns-zone/put/api/dns/v1/zones/{domain}
func (c *Client) UpdateDNSRecords(ctx context.Context, domain string, zone ZoneRequest) error {
	endpoint := c.BaseURL.JoinPath("/api/dns/v1/zones/", domain)

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, zone)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// DeleteDNSRecords deletes DNS records for the selected domain.
// https://developers.hostinger.com/#tag/dns-zone/delete/api/dns/v1/zones/{domain}
func (c *Client) DeleteDNSRecords(ctx context.Context, domain string, filters []Filter) error {
	endpoint := c.BaseURL.JoinPath("/api/dns/v1/zones/", domain)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, Filters{Filters: filters})
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set(authorizationHeader, "Bearer "+c.token)

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
