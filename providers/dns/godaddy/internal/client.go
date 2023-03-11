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

// DefaultBaseURL represents the API endpoint to call.
const DefaultBaseURL = "https://api.godaddy.com"

const authorizationHeader = "Authorization"

type Client struct {
	apiKey    string
	apiSecret string

	baseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(apiKey string, apiSecret string) *Client {
	baseURL, _ := url.Parse(DefaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) GetRecords(ctx context.Context, domainZone, rType, recordName string) ([]DNSRecord, error) {
	endpoint := c.baseURL.JoinPath("v1", "domains", domainZone, "records", rType, recordName)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var records []DNSRecord
	err = c.do(req, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (c *Client) UpdateTxtRecords(ctx context.Context, records []DNSRecord, domainZone, recordName string) error {
	endpoint := c.baseURL.JoinPath("v1", "domains", domainZone, "records", "TXT", recordName)

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, records)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set(authorizationHeader, fmt.Sprintf("sso-key %s:%s", c.apiKey, c.apiSecret))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
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
