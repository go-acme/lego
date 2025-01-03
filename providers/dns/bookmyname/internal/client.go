package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://www.bookmyname.com/dyndns/"

// Client the BookMyName API client.
type Client struct {
	username string
	password string

	baseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(username, password string) (*Client, error) {
	if username == "" || password == "" {
		return nil, errors.New("credentials missing")
	}

	return &Client{
		username:   username,
		password:   password,
		baseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddRecord(ctx context.Context, record Record) error {
	endpoint, err := c.createEndpoint(record, "add")
	if err != nil {
		return err
	}

	err = c.do(ctx, endpoint)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RemoveRecord(ctx context.Context, record Record) error {
	endpoint, err := c.createEndpoint(record, "remove")
	if err != nil {
		return err
	}

	err = c.do(ctx, endpoint)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) createEndpoint(record Record, action string) (*url.URL, error) {
	endpoint, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse URL:  %w", err)
	}

	values, err := querystring.Values(record)
	if err != nil {
		return nil, fmt.Errorf("query parameters: %w", err)
	}

	values.Set("do", action)

	endpoint.RawQuery = values.Encode()

	return endpoint, nil
}

func (c *Client) do(ctx context.Context, endpoint *url.URL) error {
	endpoint.User = url.UserPassword(c.username, c.password)

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

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	if !strings.HasPrefix(string(raw), "good: update done") && !strings.HasPrefix(string(raw), "good: remove done") {
		return fmt.Errorf("unexpected response: %s", string(bytes.TrimSpace(raw)))
	}

	return nil
}
