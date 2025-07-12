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

const defaultBaseURL = "https://api.variomedia.de"

const authorizationHeader = "Authorization"

// Client the API client for Variomedia.
type Client struct {
	apiToken string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiToken string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiToken:   apiToken,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// CreateDNSRecord creates a new DNS entry.
// https://api.variomedia.de/docs/dns-records.html#erstellen
func (c *Client) CreateDNSRecord(ctx context.Context, record DNSRecord) (*CreateDNSRecordResponse, error) {
	endpoint := c.baseURL.JoinPath("dns-records")

	data := CreateDNSRecordRequest{Data: Data{
		Type:       "dns-record",
		Attributes: record,
	}}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, data)
	if err != nil {
		return nil, err
	}

	var result CreateDNSRecordResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteDNSRecord deletes a DNS record.
// https://api.variomedia.de/docs/dns-records.html#l%C3%B6schen
func (c *Client) DeleteDNSRecord(ctx context.Context, id string) (*DeleteRecordResponse, error) {
	endpoint := c.baseURL.JoinPath("dns-records", id)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result DeleteRecordResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetJob returns a single job based on its ID.
// https://api.variomedia.de/docs/job-queue.html
func (c *Client) GetJob(ctx context.Context, id string) (*GetJobResponse, error) {
	endpoint := c.baseURL.JoinPath("queue-jobs", id)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result GetJobResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) do(req *http.Request, data any) error {
	req.Header.Set(authorizationHeader, "token "+c.apiToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return parseError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, data)
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

	req.Header.Set("Accept", "application/vnd.variomedia.v1+json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/vnd.api+json")
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

	return errAPI
}
