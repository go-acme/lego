package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	querystring "github.com/google/go-querystring/query"
)

const statusSuccess = "ok"

// Client the Technitium API client.
type Client struct {
	apiToken string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(baseURL, apiToken string) (*Client, error) {
	if apiToken == "" {
		return nil, errors.New("missing credentials")
	}

	if baseURL == "" {
		return nil, errors.New("missing server URL")
	}

	apiEndpoint, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		apiToken:   apiToken,
		baseURL:    apiEndpoint,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// AddRecord adds a resource record for an authoritative zone.
// https://github.com/TechnitiumSoftware/DnsServer/blob/master/APIDOCS.md#add-record
func (c *Client) AddRecord(ctx context.Context, record Record) (*Record, error) {
	endpoint := c.baseURL.JoinPath("api", "zones", "records", "add")

	req, err := c.newFormRequest(ctx, endpoint, record)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	result := &APIResponse[AddRecordResponse]{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	if result.Status != statusSuccess {
		return nil, result
	}

	return result.Response.AddedRecord, nil
}

// DeleteRecord deletes a record from an authoritative zone.
// https://github.com/TechnitiumSoftware/DnsServer/blob/master/APIDOCS.md#delete-record
func (c *Client) DeleteRecord(ctx context.Context, record Record) error {
	endpoint := c.baseURL.JoinPath("api", "zones", "records", "delete")

	req, err := c.newFormRequest(ctx, endpoint, record)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	result := &APIResponse[any]{}

	err = c.do(req, result)
	if err != nil {
		return err
	}

	if result.Status != statusSuccess {
		return result
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
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

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func (c *Client) newFormRequest(ctx context.Context, endpoint *url.URL, payload any) (*http.Request, error) {
	values := url.Values{}

	if payload != nil {
		var err error
		values, err = querystring.Values(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request body: %w", err)
		}
	}

	values.Set("token", c.apiToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return req, nil
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	var errAPI APIResponse[any]
	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return &errAPI
}
