package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	querystring "github.com/google/go-querystring/query"
)

// Record represents a DNS record.
type Record struct {
	Name  string `url:"name,omitempty"`
	Type  string `url:"type,omitempty"`
	Value string `url:"value,omitempty"`
	TTL   int    `url:"ttl,omitempty"`
}

// Client the Direct Admin API client.
type Client struct {
	baseURL    *url.URL
	HTTPClient *http.Client

	username string
	password string
}

// NewClient creates a new Client.
func NewClient(baseURL, username, password string) (*Client, error) {
	api, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		baseURL:    api,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		username:   username,
		password:   password,
	}, nil
}

func (c Client) SetRecord(ctx context.Context, domain string, record Record) error {
	data, err := querystring.Values(record)
	if err != nil {
		return err
	}

	data.Set("action", "add")

	return c.do(ctx, domain, data)
}

func (c Client) DeleteRecord(ctx context.Context, domain string, record Record) error {
	data, err := querystring.Values(record)
	if err != nil {
		return err
	}

	data.Set("action", "delete")

	return c.do(ctx, domain, data)
}

func (c Client) do(ctx context.Context, domain string, data url.Values) error {
	endpoint := c.baseURL.JoinPath("CMD_API_DNS_CONTROL")

	query := endpoint.Query()
	query.Set("domain", domain)
	query.Set("json", "yes")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("error: %d: %s", resp.StatusCode, string(raw))
}
