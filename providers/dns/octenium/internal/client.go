package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

	data := endpoint.Query()
	data.Set("domain-name", domain)
	endpoint.RawQuery = data.Encode()

	req, err := newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := &DomainsResponse{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Domains, nil
}

// ListDNSRecords retrieves a list of DNS records.
// https://octenium.com/api#tag/Domains-DNS/operation/domains-dns-records-list
func (c *Client) ListDNSRecords(ctx context.Context, orderID, recordType string) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", "dns-records", "list")

	data := make(url.Values)
	data.Set("order-id", orderID)
	data.Set("types[]", recordType)

	req, err := newRequest(ctx, http.MethodPost, endpoint, data)
	if err != nil {
		return nil, err
	}

	result := &ListRecordsResponse{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Records, nil
}

// AddDNSRecord adds a DNS record.
// https://octenium.com/api#tag/Domains-DNS/operation/domains-dns-records-add
func (c *Client) AddDNSRecord(ctx context.Context, orderID string, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", "dns-records", "add")

	data, err := querystring.Values(record)
	if err != nil {
		return nil, err
	}

	data.Set("order-id", orderID)

	req, err := newRequest(ctx, http.MethodPost, endpoint, data)
	if err != nil {
		return nil, err
	}

	result := &AddRecordResponse{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Record, nil
}

// DeleteDNSRecord deletes a DNS record.
// https://octenium.com/api#tag/Domains-DNS/operation/domains-dns-records-delete
func (c *Client) DeleteDNSRecord(ctx context.Context, orderID string, recordID int) (*DeletedRecordInfo, error) {
	endpoint := c.BaseURL.JoinPath("domains", "dns-records", "delete")

	data := make(url.Values)
	data.Set("order-id", orderID)
	data.Set("line", strconv.Itoa(recordID))

	req, err := newRequest(ctx, http.MethodPost, endpoint, data)
	if err != nil {
		return nil, err
	}

	result := &DeleteRecordResponse{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Deleted, nil
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

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var response APIResponse

	err = json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if response.Status != statusSuccess {
		return fmt.Errorf("unexpected status: %s: %s", response.Status, response.Error)
	}

	err = json.Unmarshal(response.Response, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, response.Response, err)
	}

	return nil
}

func newRequest(ctx context.Context, method string, endpoint *url.URL, payload url.Values) (*http.Request, error) {
	var body io.Reader = http.NoBody

	if method == http.MethodPost && payload != nil {
		body = strings.NewReader(payload.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if method == http.MethodPost && payload != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return req, nil
}
