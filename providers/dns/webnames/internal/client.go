package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://www.webnames.ru/scripts/json_domain_zone_manager.pl"

// Client the Webnames API client.
type Client struct {
	apiKey string

	baseURL    string
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) AddTXTRecord(ctx context.Context, domain, subDomain, value string) error {
	data := url.Values{}
	data.Set("domain", domain)
	data.Set("type", "TXT")
	data.Set("record", subDomain+":"+value)
	data.Set("action", "add")

	return c.doRequest(ctx, data)
}

func (c *Client) RemoveTXTRecord(ctx context.Context, domain, subDomain, value string) error {
	data := url.Values{}
	data.Set("domain", domain)
	data.Set("type", "TXT")
	data.Set("record", subDomain+":"+value)
	data.Set("action", "delete")

	return c.doRequest(ctx, data)
}

func (c *Client) doRequest(ctx context.Context, data url.Values) error {
	data.Set("apikey", c.apiKey)

	endpoint, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}

	endpoint.RawQuery = data.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var r APIResponse
	err = json.Unmarshal(raw, &r)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if r.Result == "OK" {
		return nil
	}

	return fmt.Errorf("%s: %s", r.Result, r.Details)
}
