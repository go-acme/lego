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

// DefaultBaseURL the default API endpoint.
const DefaultBaseURL = "https://api.dreamhost.com"

const (
	cmdAddRecord    = "dns-add_record"
	cmdRemoveRecord = "dns-remove_record"
)

// Client the Dreamhost API client.
type Client struct {
	apiKey string

	BaseURL    string
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		BaseURL:    DefaultBaseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// AddRecord adds a TXT record.
func (c *Client) AddRecord(ctx context.Context, domain, value string) error {
	query, err := c.buildEndpoint(cmdAddRecord, domain, value)
	if err != nil {
		return err
	}

	return c.updateTxtRecord(ctx, query)
}

// RemoveRecord removes a TXT record.
func (c *Client) RemoveRecord(ctx context.Context, domain, value string) error {
	query, err := c.buildEndpoint(cmdRemoveRecord, domain, value)
	if err != nil {
		return err
	}

	return c.updateTxtRecord(ctx, query)
}

// action is either cmdAddRecord or cmdRemoveRecord.
func (c *Client) buildEndpoint(action, domain, txt string) (*url.URL, error) {
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Set("key", c.apiKey)
	query.Set("cmd", action)
	query.Set("format", "json")
	query.Set("record", domain)
	query.Set("type", "TXT")
	query.Set("value", txt)
	query.Set("comment", url.QueryEscape("Managed By lego"))
	endpoint.RawQuery = query.Encode()

	return endpoint, nil
}

// updateTxtRecord will either add or remove a TXT record.
func (c *Client) updateTxtRecord(ctx context.Context, endpoint *url.URL) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var response apiResponse

	err = json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if response.Result == "error" {
		return fmt.Errorf("add TXT record failed: %s", response.Data)
	}

	return nil
}
