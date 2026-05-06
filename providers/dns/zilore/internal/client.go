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

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://api.zilore.com/dns/v1/"

const AccessKeyHeader = "X-Auth-Key"

// Client the Zilore API client.
type Client struct {
	accessKey string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(accessKey string) (*Client, error) {
	if accessKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		accessKey:  accessKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// AddRecord adds a TXT record.
// https://zilore.com/en/help/api/records/add_record
func (c *Client) AddRecord(ctx context.Context, domain string, record Record) (*RecordResponse, error) {
	endpoint := c.BaseURL.JoinPath("domains", domain, "records")

	req, err := newFormRequest(ctx, endpoint, record)
	if err != nil {
		return nil, err
	}

	var result APIResponse

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Response, nil
}

// DeleteRecord deletes a TXT record.
// https://zilore.com/en/help/api/records/delete_record
func (c *Client) DeleteRecord(ctx context.Context, domain string, recordID int) error {
	endpoint := c.BaseURL.JoinPath("domains", domain, "records")

	query := endpoint.Query()
	query.Set("record_id", strconv.Itoa(recordID))
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint.String(), nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set(AccessKeyHeader, c.accessKey)

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

func newFormRequest(ctx context.Context, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		values, err := querystring.Values(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request body: %w", err)
		}

		buf = bytes.NewBufferString(values.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
