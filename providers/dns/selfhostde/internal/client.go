package internal

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://selfhost.de/cgi-bin/api.pl"

// Client the SelfHost client.
type Client struct {
	username string
	password string

	baseURL    string
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(username, password string) *Client {
	return &Client{
		username:   username,
		password:   password,
		baseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// UpdateTXTRecord updates content of an existing TXT record.
func (c *Client) UpdateTXTRecord(ctx context.Context, recordID, content string) error {
	endpoint, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("parse URL:  %w", err)
	}

	query := endpoint.Query()
	query.Set("username", c.username)
	query.Set("password", c.password)
	query.Set("rid", recordID)
	query.Set("content", content)

	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("new HTTP request: %w", err)
	}

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
