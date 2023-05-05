package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// defaultBaseURL represents the API endpoint to call.
const defaultBaseURL = "https://napi.arvancloud.ir"

const authorizationHeader = "Authorization"

// Client the ArvanCloud client.
type Client struct {
	apiKey string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(apiKey string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetTxtRecord gets a TXT record.
func (c *Client) GetTxtRecord(ctx context.Context, domain, name, value string) (*DNSRecord, error) {
	records, err := c.getRecords(ctx, domain, name)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if equalsTXTRecord(record, name, value) {
			return &record, nil
		}
	}

	return nil, fmt.Errorf("could not find record: Domain: %s; Record: %s", domain, name)
}

// https://www.arvancloud.ir/docs/api/cdn/4.0#operation/dns_records.list
func (c *Client) getRecords(ctx context.Context, domain, search string) ([]DNSRecord, error) {
	endpoint := c.baseURL.JoinPath("cdn", "4.0", "domains", domain, "dns-records")

	if search != "" {
		query := endpoint.Query()
		query.Set("search", strings.ReplaceAll(search, "_", ""))
		endpoint.RawQuery = query.Encode()
	}

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	response := &apiResponse[[]DNSRecord]{}
	err = c.do(req, http.StatusOK, response)
	if err != nil {
		return nil, fmt.Errorf("could not get records %s: Domain: %s: %w", search, domain, err)
	}

	return response.Data, nil
}

// CreateRecord creates a DNS record.
// https://www.arvancloud.ir/docs/api/cdn/4.0#operation/dns_records.create
func (c *Client) CreateRecord(ctx context.Context, domain string, record DNSRecord) (*DNSRecord, error) {
	endpoint := c.baseURL.JoinPath("cdn", "4.0", "domains", domain, "dns-records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	response := &apiResponse[*DNSRecord]{}
	err = c.do(req, http.StatusCreated, response)
	if err != nil {
		return nil, fmt.Errorf("could not create record; Domain: %s: %w", domain, err)
	}

	return response.Data, nil
}

// DeleteRecord deletes a DNS record.
// https://www.arvancloud.ir/docs/api/cdn/4.0#operation/dns_records.remove
func (c *Client) DeleteRecord(ctx context.Context, domain, id string) error {
	endpoint := c.baseURL.JoinPath("cdn", "4.0", "domains", domain, "dns-records", id)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	err = c.do(req, http.StatusOK, nil)
	if err != nil {
		return fmt.Errorf("could not delete record %s; Domain: %s: %w", id, domain, err)
	}

	return nil
}

func (c *Client) do(req *http.Request, expectedStatus int, result any) error {
	req.Header.Set(authorizationHeader, c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != expectedStatus {
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

func equalsTXTRecord(record DNSRecord, name, value string) bool {
	if record.Type != "txt" {
		return false
	}

	if record.Name != name {
		return false
	}

	data, ok := record.Value.(map[string]any)
	if !ok {
		return false
	}

	return data["text"] == value
}
