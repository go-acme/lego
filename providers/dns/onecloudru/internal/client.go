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

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

const defaultBaseURL = "https://api.1cloud.ru"

// Client the 1cloud.ru API client.
type Client struct {
	token string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		token:      token,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// GetDomains returns the list of domains.
// https://1cloud.ru/api/dns/domainlist
func (c *Client) GetDomains(ctx context.Context) ([]Domain, error) {
	endpoint := c.BaseURL.JoinPath("dns")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result []Domain

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateTXTRecord creates a TXT record.
// https://1cloud.ru/api/dns/createrecordtxt
func (c *Client) CreateTXTRecord(ctx context.Context, request CreateTXTRecordRequest) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("dns", "recordtxt")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, request)
	if err != nil {
		return nil, err
	}

	result := new(Record)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteRecord deletes a TXT record.
// https://1cloud.ru/api/dns/deleterecord
func (c *Client) DeleteRecord(ctx context.Context, domainID, recordID int64) error {
	endpoint := c.BaseURL.JoinPath("dns", strconv.FormatInt(domainID, 10), strconv.FormatInt(recordID, 10))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
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

// TTLRounder Rounds the given TTL in seconds to the next accepted value.
func TTLRounder(ttl int) int {
	for _, validTTL := range []int{1, 5, 30, 60, 300, 600, 900, 1800, 3600, 7200, 21160, 43200, 86400} {
		if ttl <= validTTL {
			return validTTL
		}
	}

	return 2592000
}
