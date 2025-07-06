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

const defaultBaseURL = "https://mijn.host/api/v2/"

const authorizationHeader = "API-Key"

// Client a mijn.host DNS API client.
type Client struct {
	apiKey string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ListDomains Retrieve all domains from an account.
// https://mijn.host/api/doc/api-3563872
func (c *Client) ListDomains(ctx context.Context) ([]Domain, error) {
	endpoint := c.baseURL.JoinPath("domains")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	var results Response[DomainData]
	err = c.do(req, &results)
	if err != nil {
		return nil, err
	}

	return results.Data.Domains, nil
}

// GetRecords Retrieve DNS records of specific domain.
// https://mijn.host/api/doc/api-3563906
func (c *Client) GetRecords(ctx context.Context, domain string) ([]Record, error) {
	endpoint := c.baseURL.JoinPath("domains", domain, "dns")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	var results Response[RecordData]
	err = c.do(req, &results)
	if err != nil {
		return nil, err
	}

	return results.Data.Records, nil
}

// UpdateRecords Update DNS records of specific domain.
// https://mijn.host/api/doc/api-3563907
func (c *Client) UpdateRecords(ctx context.Context, domain string, records []Record) error {
	endpoint := c.baseURL.JoinPath("domains", domain, "dns")

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, RecordData{Records: records})
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set(authorizationHeader, c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
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
