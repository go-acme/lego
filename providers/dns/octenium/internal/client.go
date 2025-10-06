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
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://api.panel.octenium.com/"

const statusSuccess = "success"

// Client the Octenium API client.
type Client struct {
	apiKey string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// ListDomains retrieves a list of domains.
// https://octenium.com/api#tag/Domains/operation/listdomains
func (c *Client) ListDomains(ctx context.Context, domain string) (map[string]Domain, error) {
	endpoint := c.BaseURL.JoinPath("domains")

	query := endpoint.Query()
	query.Set("domain-name", domain)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}

	var result APIResponse[[]DomainsResponse]

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	if result.Status != statusSuccess {
		return nil, fmt.Errorf("unexpected status: %s", result.Status)
	}

	allDomains := make(map[string]Domain)
	for _, domains := range result.Response {
		for k, v := range domains.Domains {
			allDomains[k] = v
		}
	}

	return allDomains, nil
}

// ListDNSRecords retrieves a list of DNS records.
// https://octenium.com/api#tag/Domains-DNS/operation/domains-dns-records-list
func (c *Client) ListDNSRecords(ctx context.Context, orderID string) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", "dns-records", "list")

	query := endpoint.Query()
	query.Set("order-id", orderID)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint)
	if err != nil {
		return nil, err
	}

	var result APIResponse[ListRecordsResponse]

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	if result.Status != statusSuccess {
		return nil, fmt.Errorf("unexpected status: %s", result.Status)
	}

	return result.Response.Records, nil
}

// AddDNSRecord adds a DNS record.
// https://octenium.com/api#tag/Domains-DNS/operation/domains-dns-records-add
func (c *Client) AddDNSRecord(ctx context.Context, orderID string, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", "dns-records", "add")

	values, err := querystring.Values(record)
	if err != nil {
		return nil, err
	}

	values.Set("order-id", orderID)
	endpoint.RawQuery = values.Encode()

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint)
	if err != nil {
		return nil, err
	}

	var result APIResponse[AddRecordResponse]

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	if result.Status != statusSuccess {
		return nil, fmt.Errorf("unexpected status: %s", result.Status)
	}

	return result.Response.Record, nil
}

// DeleteDNSRecord deletes a DNS record.
// https://octenium.com/api#tag/Domains-DNS/operation/domains-dns-records-delete
func (c *Client) DeleteDNSRecord(ctx context.Context, orderID string, recordID int) (*DeletedRecordInfo, error) {
	endpoint := c.BaseURL.JoinPath("domains", "dns-records", "delete")

	query := endpoint.Query()
	query.Set("order-id", orderID)
	query.Set("line", strconv.Itoa(recordID))
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint)
	if err != nil {
		return nil, err
	}

	var result APIResponse[DeleteRecordResponse]
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	if result.Status != statusSuccess {
		return nil, fmt.Errorf("unexpected status: %s", result.Status)
	}

	return result.Response.Deleted, nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set("X-Api-Key", c.apiKey)

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

func newJSONRequest(ctx context.Context, method string, endpoint *url.URL) (*http.Request, error) {
	buf := new(bytes.Buffer)

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}
