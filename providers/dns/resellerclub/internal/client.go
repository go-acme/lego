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

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

const defaultBaseURL = "https://api.sandbox.resellerclub.com/v2"

const UserIDHeader = "X-User-Id"

// Client the ResellerClub API client.
type Client struct {
	apiKey string
	userID string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey, userID string) (*Client, error) {
	if apiKey == "" || userID == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		userID:     userID,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateDNSRecord adds a new DNS record to a domain's zone.
// https://developer.resellerclub.com/dns/apis#create-a-dns-record
func (c *Client) CreateDNSRecord(ctx context.Context, domainName string, record Record) (*APIResponse, error) {
	endpoint := c.BaseURL.JoinPath("dns", "zones", domainName, "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteDNSRecord removes a specific DNS record from a domain's zone.
// https://developer.resellerclub.com/dns/apis#delete-a-dns-record
func (c *Client) DeleteDNSRecord(ctx context.Context, domainName string, record Record) (*APIResponse, error) {
	endpoint := c.BaseURL.JoinPath("dns", "zones", domainName, "records")

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	// https://developer.resellerclub.com/documentation/authentication
	req.Header.Set("Authorization", "ApiKey "+c.apiKey)

	req.Header.Set(UserIDHeader, c.userID)

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
