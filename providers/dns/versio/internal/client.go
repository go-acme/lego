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

// DefaultBaseURL default API endpoint.
const DefaultBaseURL = "https://www.versio.nl/api/v1/"

// Client the API client for Versio DNS.
type Client struct {
	username string
	password string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(username string, password string) *Client {
	baseURL, _ := url.Parse(DefaultBaseURL)

	return &Client{
		username:   username,
		password:   password,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// UpdateDomain updates domain information.
// https://www.versio.nl/RESTapidoc/#api-Domains-Update
func (c *Client) UpdateDomain(ctx context.Context, domain string, msg *DomainInfo) (*DomainInfoResponse, error) {
	endpoint := c.BaseURL.JoinPath("domains", domain, "update")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, msg)
	if err != nil {
		return nil, err
	}

	respData := &DomainInfoResponse{}
	err = c.do(req, respData)
	if err != nil {
		return nil, err
	}

	return respData, nil
}

// GetDomain gets domain information.
// https://www.versio.nl/RESTapidoc/#api-Domains-Domain
func (c *Client) GetDomain(ctx context.Context, domain string) (*DomainInfoResponse, error) {
	endpoint := c.BaseURL.JoinPath("domains", domain)

	query := endpoint.Query()
	query.Set("show_dns_records", "true")
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	respData := &DomainInfoResponse{}
	err = c.do(req, respData)
	if err != nil {
		return nil, err
	}

	return respData, nil
}

func (c *Client) do(req *http.Request, result any) error {
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.HTTPClient.Do(req)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

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

	if err = json.Unmarshal(raw, result); err != nil {
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

	response := &ErrorResponse{}
	err := json.Unmarshal(raw, response)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code: %d] %w", resp.StatusCode, response.Message)
}
