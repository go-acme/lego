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
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://api.simply.com/2/"

// Client is a Simply.com API client.
type Client struct {
	accountName string
	apiKey      string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(accountName, apiKey string) (*Client, error) {
	if accountName == "" {
		return nil, errors.New("credentials missing: accountName")
	}

	if apiKey == "" {
		return nil, errors.New("credentials missing: apiKey")
	}

	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		accountName: accountName,
		apiKey:      apiKey,
		baseURL:     baseURL,
		HTTPClient:  &http.Client{Timeout: 5 * time.Second},
	}, nil
}

// GetRecords lists all the records in the zone.
func (c *Client) GetRecords(ctx context.Context, zoneName string) ([]Record, error) {
	endpoint := c.createEndpoint(zoneName, "/")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	result := &apiResponse[[]Record, json.RawMessage]{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Records, nil
}

// AddRecord adds a record.
func (c *Client) AddRecord(ctx context.Context, zoneName string, record Record) (int64, error) {
	endpoint := c.createEndpoint(zoneName, "/")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	result := &apiResponse[json.RawMessage, recordHeader]{}

	err = c.do(req, result)
	if err != nil {
		return 0, err
	}

	return result.Record.ID, nil
}

// EditRecord updates a record.
func (c *Client) EditRecord(ctx context.Context, zoneName string, id int64, record Record) error {
	endpoint := c.createEndpoint(zoneName, strconv.FormatInt(id, 10))

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, record)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.do(req, &apiResponse[json.RawMessage, json.RawMessage]{})
}

// DeleteRecord deletes a record.
func (c *Client) DeleteRecord(ctx context.Context, zoneName string, id int64) error {
	endpoint := c.createEndpoint(zoneName, strconv.FormatInt(id, 10))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.do(req, &apiResponse[json.RawMessage, json.RawMessage]{})
}

func (c *Client) createEndpoint(zoneName, uri string) *url.URL {
	return c.baseURL.JoinPath("my", "products", zoneName, "dns", "records", strings.TrimSuffix(uri, "/"))
}

func (c *Client) do(req *http.Request, result Response) error {
	req.SetBasicAuth(c.accountName, c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusInternalServerError {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if result.GetStatus() != http.StatusOK {
		return fmt.Errorf("unexpected error: %s", result.GetMessage())
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
