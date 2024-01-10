package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://www.webnames.ru/scripts/json_domain_zone_manager.pl"

// Client the Webnames API client.
type Client struct {
	token string

	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c Client) AddTXTRecord(ctx context.Context, domain, subDomain, value string) error {
	endpoint, _ := url.Parse(defaultBaseURL)

	query := endpoint.Query()
	query.Set("apikey", c.token)
	query.Set("domain", domain)
	query.Set("type", "TXT")
	query.Set("record", strings.Join([]string{subDomain, value}, ":"))
	query.Set("action", "add")

	endpoint.RawQuery = query.Encode()

	err := c.doRequest(ctx, endpoint)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) RemoveTXTRecord(ctx context.Context, domain, subDomain, value string) error {
	endpoint, _ := url.Parse(defaultBaseURL)

	query := endpoint.Query()
	query.Set("apikey", c.token)
	query.Set("domain", domain)
	query.Set("type", "TXT")
	query.Set("record", strings.Join([]string{subDomain, value}, ":"))
	query.Set("action", "delete")

	endpoint.RawQuery = query.Encode()

	err := c.doRequest(ctx, endpoint)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) doRequest(ctx context.Context, endpoint *url.URL) error {
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

	//body := string(raw)
	var body map[string]any
	err = json.Unmarshal(raw, &body)
	if err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}

	if body["result"].(string) != "OK" {
		return fmt.Errorf("request to change TXT record for Webnames returned the following result (%s) this does not match expectation (OK) used url [%s]", body["result"].(string), endpoint)
	}

	return nil
}
