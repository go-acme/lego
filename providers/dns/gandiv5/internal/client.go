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

	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// defaultBaseURL endpoint is the Gandi API endpoint used by Present and CleanUp.
const defaultBaseURL = "https://dns.api.gandi.net/api/v5"

// APIKeyHeader API key header.
const APIKeyHeader = "X-Api-Key"

// Client the Gandi API v5 client.
type Client struct {
	apiKey string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(apiKey string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) AddTXTRecord(ctx context.Context, domain, name, value string, ttl int) error {
	// Get exiting values for the TXT records
	// Needed to create challenges for both wildcard and base name domains
	txtRecord, err := c.getTXTRecord(ctx, domain, name)
	if err != nil {
		return err
	}

	values := []string{value}
	if len(txtRecord.RRSetValues) > 0 {
		values = append(values, txtRecord.RRSetValues...)
	}

	newRecord := &Record{RRSetTTL: ttl, RRSetValues: values}

	err = c.addTXTRecord(ctx, domain, name, newRecord)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) getTXTRecord(ctx context.Context, domain, name string) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", domain, "records", name, "TXT")

	// Get exiting values for the TXT records
	// Needed to create challenges for both wildcard and base name domains
	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	txtRecord := &Record{}
	err = c.do(req, txtRecord)
	if err != nil {
		return nil, fmt.Errorf("unable to get TXT records for domain %s and name %s: %w", domain, name, err)
	}

	return txtRecord, nil
}

func (c *Client) addTXTRecord(ctx context.Context, domain, name string, newRecord *Record) error {
	endpoint := c.BaseURL.JoinPath("domains", domain, "records", name, "TXT")

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, newRecord)
	if err != nil {
		return err
	}

	message := apiResponse{}
	err = c.do(req, &message)
	if err != nil {
		return fmt.Errorf("unable to create TXT record for domain %s and name %s: %w", domain, name, err)
	}

	if message.Message != "" {
		log.Infof("API response: %s", message.Message)
	}

	return nil
}

func (c *Client) DeleteTXTRecord(ctx context.Context, domain, name string) error {
	endpoint := c.BaseURL.JoinPath("domains", domain, "records", name, "TXT")

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	message := apiResponse{}
	err = c.do(req, &message)
	if err != nil {
		return fmt.Errorf("unable to delete TXT record for domain %s and name %s: %w", domain, name, err)
	}

	if message.Message != "" {
		log.Infof("API response: %s", message.Message)
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	if c.apiKey != "" {
		req.Header.Set(APIKeyHeader, c.apiKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	err = checkResponse(req, resp)
	if err != nil {
		return err
	}

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if len(raw) > 0 {
		err = json.Unmarshal(raw, result)
		if err != nil {
			return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
		}
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

func checkResponse(req *http.Request, resp *http.Response) error {
	if resp.StatusCode == http.StatusNotFound && resp.Request.Method == http.MethodGet {
		return nil
	}

	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	return parseError(req, resp)
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	response := apiResponse{}
	err := json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("%d: request failed: %s", resp.StatusCode, response.Message)
}
