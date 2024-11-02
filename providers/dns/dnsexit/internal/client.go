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
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

const defaultBaseURL = "https://api.dnsexit.com/dns/"

// Client the DNSExit API client.
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

// AddRecord adds a record.
// https://dnsexit.com/dns/dns-api/#example-add-spf
// https://dnsexit.com/dns/dns-api/#example-lse
func (c *Client) AddRecord(ctx context.Context, domain string, record Record) error {
	payload := APIRequest{
		Domain: domain,
		Add:    []Record{record},
	}

	req, err := newJSONRequest(ctx, http.MethodPost, c.BaseURL, payload)
	if err != nil {
		return err
	}

	err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

// DeleteRecord deletes a record.
// https://dnsexit.com/dns/dns-api/#delete-a-record
func (c *Client) DeleteRecord(ctx context.Context, domain string, record Record) error {
	payload := APIRequest{
		Domain: domain,
		Delete: []Record{record},
	}

	req, err := newJSONRequest(ctx, http.MethodPost, c.BaseURL, payload)
	if err != nil {
		return err
	}

	err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) do(req *http.Request) error {
	useragent.SetHeader(req.Header)

	req.Header.Set("apikey", c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode > http.StatusBadRequest {
		return parseError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	result := &APIResponse{}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if result.Code != 0 {
		return result
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

	var errAPI APIResponse

	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return &errAPI
}
