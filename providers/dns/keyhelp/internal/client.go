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
)

// APIKeyHeader API key header.
const APIKeyHeader = "X-Api-Key"

// Client the KeyHelp API client.
type Client struct {
	apiKey string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(baseURL, apiKey string) (*Client, error) {
	if baseURL == "" {
		return nil, errors.New("missing base URL")
	}

	if apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL:  %w", err)
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    base.JoinPath("api", "v2"),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set(APIKeyHeader, c.apiKey)

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

func (c *Client) ListDomains(ctx context.Context) ([]Domain, error) {
	endpoint := c.baseURL.JoinPath("domains")

	query := endpoint.Query()
	query.Set("sort", "domain_utf8")
	endpoint.RawQuery = query.Encode()

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

func (c *Client) ListDomainRecords(ctx context.Context, domainID int) (*DomainRecords, error) {
	endpoint := c.baseURL.JoinPath("dns", strconv.Itoa(domainID))

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result DomainRecords

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) UpdateDomainRecords(ctx context.Context, domainID int, records DomainRecords) (*DomainID, error) {
	endpoint := c.baseURL.JoinPath("dns", strconv.Itoa(domainID))

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, records)
	if err != nil {
		return nil, err
	}

	var result DomainID

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
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
