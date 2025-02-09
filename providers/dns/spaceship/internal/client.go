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

const defaultBaseURL = "https://spaceship.dev/api/v1/"

// Client the Spaceship API client.
type Client struct {
	apiKey    string
	apiSecret string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey, apiSecret string) (*Client, error) {
	if apiKey == "" || apiSecret == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Add("X-Api-Secret", c.apiSecret)
	req.Header.Add("X-Api-Key", c.apiKey)

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

func (c *Client) AddRecord(ctx context.Context, domain string, record Record) error {
	endpoint := c.baseURL.JoinPath("dns", "records", domain)

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, Foo{Items: []Record{record}})
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteRecord(ctx context.Context, domain string, record Record) error {
	endpoint := c.baseURL.JoinPath("dns", "records", domain)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, []Record{record})
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetRecords(ctx context.Context, domain string) ([]Record, error) {
	endpoint := c.baseURL.JoinPath("dns", "records", domain)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result GetRecordsResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
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
