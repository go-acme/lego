package internal

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://www.mydns.jp/directedit.html"

// Client the MyDNS.jp client.
type Client struct {
	masterID string
	password string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(masterID, password string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		masterID:   masterID,
		password:   password,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) AddTXTRecord(ctx context.Context, domain, value string) error {
	return c.doRequest(ctx, domain, value, "REGIST")
}

func (c *Client) DeleteTXTRecord(ctx context.Context, domain, value string) error {
	return c.doRequest(ctx, domain, value, "DELETE")
}

func (c *Client) buildRequest(ctx context.Context, domain, value, cmd string) (*http.Request, error) {
	params := url.Values{}
	params.Set("CERTBOT_DOMAIN", domain)
	params.Set("CERTBOT_VALIDATION", value)
	params.Set("EDIT_CMD", cmd)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL.String(), strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func (c *Client) doRequest(ctx context.Context, domain, value, cmd string) error {
	req, err := c.buildRequest(ctx, domain, value, cmd)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.masterID, c.password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	return nil
}
