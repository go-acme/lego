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

const defaultBaseURL = "https://api.mittwald.de/v2/"

const authorizationHeader = "Authorization"

// Client the Mittwald client.
type Client struct {
	token string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(token string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		token:      token,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// ListDomains List Domains.
// https://api.mittwald.de/v2/docs/#/Domain/domain-list-domains
func (c Client) ListDomains(ctx context.Context) ([]Domain, error) {
	endpoint := c.baseURL.JoinPath("domains")

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

// GetDNSZone Get a DNSZone.
// https://api.mittwald.de/v2/docs/#/Domain/dns-get-dns-zone
func (c Client) GetDNSZone(ctx context.Context, zoneID string) (*DNSZone, error) {
	endpoint := c.baseURL.JoinPath("dns-zones", zoneID)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := &DNSZone{}
	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListDNSZones List DNSZones belonging to a Project.
// https://api.mittwald.de/v2/docs/#/Domain/dns-list-dns-zones
func (c Client) ListDNSZones(ctx context.Context, projectID string) ([]DNSZone, error) {
	endpoint := c.baseURL.JoinPath("projects", projectID, "dns-zones")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result []DNSZone
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateDNSZone Create a DNSZone.
// https://api.mittwald.de/v2/docs/#/Domain/dns-create-dns-zone
func (c Client) CreateDNSZone(ctx context.Context, zone CreateDNSZoneRequest) (*DNSZone, error) {
	endpoint := c.baseURL.JoinPath("dns-zones")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, zone)
	if err != nil {
		return nil, err
	}

	result := &DNSZone{}
	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateTXTRecord Update a record set on a DNSZone.
// https://api.mittwald.de/v2/docs/#/Domain/dns-update-record-set
func (c Client) UpdateTXTRecord(ctx context.Context, zoneID string, record TXTRecord) error {
	endpoint := c.baseURL.JoinPath("dns-zones", zoneID, "record-sets", "txt")

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// DeleteDNSZone Delete a DNSZone.
// https://api.mittwald.de/v2/docs/#/Domain/dns-delete-dns-zone
func (c Client) DeleteDNSZone(ctx context.Context, zoneID string) error {
	endpoint := c.baseURL.JoinPath("dns-zones", zoneID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c Client) do(req *http.Request, result any) error {
	req.Header.Set(authorizationHeader, "Bearer "+c.token)

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

	var response APIError
	err := json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code %d] %w", resp.StatusCode, response)
}
