package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL string = "https://api.domeneshop.no/v0"

// Client implements a very simple wrapper around the Domeneshop API.
// For now, it will only deal with adding and removing TXT records, as required by ACME providers.
// https://api.domeneshop.no/docs/
type Client struct {
	apiToken  string
	apiSecret string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient returns an instance of the Domeneshop API wrapper.
func NewClient(apiToken, apiSecret string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiToken:   apiToken,
		apiSecret:  apiSecret,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetDomainByName fetches the domain list and returns the Domain object for the matching domain.
// https://api.domeneshop.no/docs/#operation/getDomains
func (c *Client) GetDomainByName(ctx context.Context, domain string) (*Domain, error) {
	endpoint := c.baseURL.JoinPath("domains")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var domains []Domain

	err = c.do(req, &domains)
	if err != nil {
		return nil, err
	}

	for _, d := range domains {
		if !d.Services.DNS {
			// Domains without DNS service cannot have DNS record added.
			continue
		}

		if d.Name == domain {
			return &d, nil
		}
	}

	return nil, fmt.Errorf("failed to find matching domain name: %s", domain)
}

// CreateTXTRecord creates a TXT record with the provided host (subdomain) and data.
// https://api.domeneshop.no/docs/#tag/dns/paths/~1domains~1{domainId}~1dns/post
func (c *Client) CreateTXTRecord(ctx context.Context, domain *Domain, host string, data string) error {
	endpoint := c.baseURL.JoinPath("domains", strconv.Itoa(domain.ID), "dns")

	record := DNSRecord{
		Data: data,
		Host: host,
		TTL:  300,
		Type: "TXT",
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// DeleteTXTRecord deletes the DNS record matching the provided host and data.
// https://api.domeneshop.no/docs/#tag/dns/paths/~1domains~1{domainId}~1dns~1{recordId}/delete
func (c *Client) DeleteTXTRecord(ctx context.Context, domain *Domain, host string, data string) error {
	record, err := c.getDNSRecordByHostData(ctx, *domain, host, data)
	if err != nil {
		return err
	}

	endpoint := c.baseURL.JoinPath("domains", strconv.Itoa(domain.ID), "dns", strconv.Itoa(record.ID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// getDNSRecordByHostData finds the first matching DNS record with the provided host and data.
// https://api.domeneshop.no/docs/#operation/getDnsRecords
func (c *Client) getDNSRecordByHostData(ctx context.Context, domain Domain, host string, data string) (*DNSRecord, error) {
	endpoint := c.baseURL.JoinPath("domains", strconv.Itoa(domain.ID), "dns")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var records []DNSRecord

	err = c.do(req, &records)
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		if r.Host == host && r.Data == data {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("failed to find record with host %s for domain %s", host, domain.Name)
}

// do a request against the API,
// and makes sure that the required Authorization header is set using `setBasicAuth`.
func (c *Client) do(req *http.Request, result any) error {
	req.SetBasicAuth(c.apiToken, c.apiSecret)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
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
