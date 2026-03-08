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

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://rest-api.eurodns.com/dns-zones/"

const (
	HeaderAppID  = "X-APP-ID"
	HeaderAPIKey = "X-API-KEY"
)

// Client the EuroDNS API client.
type Client struct {
	appID  string
	apiKey string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(appID, apiKey string) (*Client, error) {
	if appID == "" || apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		appID:      appID,
		apiKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// GetZone gets a DNS Zone.
// https://docapi.eurodns.com/#/dnsprovider/getdnszone
func (c *Client) GetZone(ctx context.Context, domain string) (*Zone, error) {
	endpoint := c.BaseURL.JoinPath(domain)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := &Zone{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// SaveZone saves a DNS Zone.
// https://docapi.eurodns.com/#/dnsprovider/savednszone
func (c *Client) SaveZone(ctx context.Context, domain string, zone *Zone) error {
	endpoint := c.BaseURL.JoinPath(domain)

	if len(zone.URLForwards) == 0 {
		zone.URLForwards = make([]URLForward, 0)
	}

	if len(zone.MailForwards) == 0 {
		zone.MailForwards = make([]MailForward, 0)
	}

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, zone)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// ValidateZone validates DNS Zone.
// https://docapi.eurodns.com/#/dnsprovider/checkdnszone
func (c *Client) ValidateZone(ctx context.Context, domain string, zone *Zone) (*Zone, error) {
	endpoint := c.BaseURL.JoinPath(domain, "check")

	if len(zone.URLForwards) == 0 {
		zone.URLForwards = make([]URLForward, 0)
	}

	if len(zone.MailForwards) == 0 {
		zone.MailForwards = make([]MailForward, 0)
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, zone)
	if err != nil {
		return nil, err
	}

	result := &Zone{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set(HeaderAppID, c.appID)
	req.Header.Set(HeaderAPIKey, c.apiKey)

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

	return fmt.Errorf("%d: %w", resp.StatusCode, &errAPI)
}

const DefaultTTL = 600

// TTLRounder rounds the given TTL in seconds to the next accepted value.
// Accepted TTL values are: 600, 900, 1800,3600, 7200, 14400, 21600, 43200, 86400, 172800, 432000, 604800.
func TTLRounder(ttl int) int {
	for _, validTTL := range []int{DefaultTTL, 900, 1800, 3600, 7200, 14400, 21600, 43200, 86400, 172800, 432000, 604800} {
		if ttl <= validTTL {
			return validTTL
		}
	}

	return DefaultTTL
}
