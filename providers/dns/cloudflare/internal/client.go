/*
Package internal Cloudflare API client.

The official client is huge and still growing.
- https://github.com/cloudflare/cloudflare-go/issues/4171
*/
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

	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

const defaultBaseURL = "https://api.cloudflare.com/client/v4"

// Client the Cloudflare API client.
type Client struct {
	authEmail string
	authKey   string
	authToken string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(opts ...Option) (*Client, error) {
	baseURL, _ := url.Parse(defaultBaseURL)

	client := &Client{
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}

	for _, opt := range opts {
		err := opt(client)
		if err != nil {
			return nil, err
		}
	}

	if client.authToken != "" {
		return client, nil
	}

	if client.authEmail == "" && client.authKey == "" {
		return nil, errors.New("invalid credentials: authEmail, authKey or authToken must be set")
	}

	if client.authEmail == "" || client.authKey == "" {
		return nil, errors.New("invalid credentials: authEmail and authKey must be set together")
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return client, nil
}

// CreateDNSRecord creates a new DNS record for a zone.
// https://developers.cloudflare.com/api/resources/dns/subresources/records/methods/create/
func (c *Client) CreateDNSRecord(ctx context.Context, zoneID string, record Record) (*Record, error) {
	endpoint := c.baseURL.JoinPath("zones", zoneID, "dns_records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	var result APIResponse[Record]

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result.Result, nil
}

// DeleteDNSRecord deletes DNS record.
// https://developers.cloudflare.com/api/resources/dns/subresources/records/methods/delete/
func (c *Client) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	endpoint := c.baseURL.JoinPath("zones", zoneID, "dns_records", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// ZonesByName returns a list of zones matching the given name.
// https://developers.cloudflare.com/api/resources/zones/methods/list/
func (c *Client) ZonesByName(ctx context.Context, name string) ([]Zone, error) {
	endpoint := c.baseURL.JoinPath("zones")

	query := endpoint.Query()
	query.Set("name", name)
	query.Set("per_page", "50")
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result APIResponse[[]Zone]

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Result, nil
}

func (c *Client) do(req *http.Request, result any) error {
	// https://developers.cloudflare.com/fundamentals/api/how-to/make-api-calls/
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	} else {
		req.Header.Set("X-Auth-Email", c.authEmail)
		req.Header.Set("X-Auth-Key", c.authKey)
	}

	useragent.SetHeader(req.Header)

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

	var response APIResponse[any]

	err := json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code %d] %w", resp.StatusCode, response.Errors)
}
