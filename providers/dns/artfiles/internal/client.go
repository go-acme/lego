package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

const defaultBaseURL = "https://dcp.c.artfiles.de/api/"

// Client the ArtFiles API client.
type Client struct {
	username string
	password string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(username, password string) (*Client, error) {
	if username == "" || password == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		username:   username,
		password:   password,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) GetDomains(ctx context.Context) ([]string, error) {
	endpoint := c.BaseURL.JoinPath("domain", "get_domains.html")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	raw, err := c.do(req)
	if err != nil {
		return nil, err
	}

	return parseDomains(string(raw))
}

func (c *Client) GetRecords(ctx context.Context, domain string) (map[string]json.RawMessage, error) {
	endpoint := c.BaseURL.JoinPath("dns", "get_dns.html")

	query := endpoint.Query()
	query.Set("domain", domain)

	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	raw, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var result Records

	err = json.Unmarshal(raw, &result)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, http.StatusOK, raw, err)
	}

	return result.Data, nil
}

func (c *Client) SetRecords(ctx context.Context, domain, rType string, value RecordValue) error {
	endpoint := c.BaseURL.JoinPath("dns", "set_dns.html")

	query := endpoint.Query()
	query.Set("domain", domain)
	query.Set(rType, value.String())

	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	_, err = c.do(req)

	return err
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	useragent.SetHeader(req.Header)

	req.SetBasicAuth(c.username, c.password)

	if req.Method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return raw, nil
}
