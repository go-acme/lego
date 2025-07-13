/*
Package internal Civo API client.

Because the dependencies on k8s, the official client cannot be used.
- https://github.com/civo/civogo/blob/v0.2.99/go.mod -> k8s.io/client-go
- https://github.com/civo/civogo/blob/v0.3.34/go.mod -> k8s.io/api
- https://github.com/civo/civogo/blob/v0.3.38/go.mod -> k8s.io/api + k8s.io/apimachinery
- Current version -> https://github.com/civo/civogo/blob/v0.6.1/go.mod
*/
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
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	"golang.org/x/oauth2"
)

const defaultBaseURL = "https://api.civo.com/v2"

// Client the Civo API client.
type Client struct {
	region string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(hc *http.Client, region string) (*Client, error) {
	baseURL, _ := url.Parse(defaultBaseURL)

	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{
		region:     region,
		BaseURL:    baseURL,
		HTTPClient: hc,
	}, nil
}

// ListDomains a list of all domain names within the account.
// https://www.civo.com/api/dns#list-domain-names
func (c *Client) ListDomains(ctx context.Context) ([]Domain, error) {
	endpoint := c.BaseURL.JoinPath("dns")

	req, err := c.newJSONRequest(ctx, http.MethodGet, endpoint, nil)
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

// ListDNSRecords a list of all DNS records in the specified domain.
// https://www.civo.com/api/dns#list-dns-records
func (c *Client) ListDNSRecords(ctx context.Context, domainID string) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("dns", domainID, "records")

	req, err := c.newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result []Record

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateDNSRecord creates DNS records for a specific domain.
// https://www.civo.com/api/dns#create-a-new-dns-record
func (c *Client) CreateDNSRecord(ctx context.Context, domainID string, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("dns", domainID, "records")

	req, err := c.newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	var result Record

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteDNSRecord remove a DNS record from a domain.
// https://www.civo.com/api/dns#deleting-a-dns-record
func (c *Client) DeleteDNSRecord(ctx context.Context, record Record) error {
	endpoint := c.BaseURL.JoinPath("dns", record.DomainID, "records", record.ID)

	req, err := c.newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
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

func (c *Client) newJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	if method == http.MethodGet || method == http.MethodDelete {
		query := endpoint.Query()
		query.Set("region", c.region)

		endpoint.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	useragent.SetHeader(req.Header)

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

// OAuthStaticAccessToken Authorization header.
// https://www.civo.com/api#authentication
func OAuthStaticAccessToken(client *http.Client, accessToken string) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	client.Transport = &oauth2.Transport{
		Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
		Base:   client.Transport,
	}

	return client
}
