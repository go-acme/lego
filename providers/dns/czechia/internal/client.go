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
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

const defaultBaseURL = "https://api.czechia.com/api"

const authorizationTokenHeader = "AuthorizationToken"

// Client the Czechia API client.
type Client struct {
	token string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		token:      token,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddTXTRecord(ctx context.Context, domain string, record TXTRecord) error {
	endpoint := c.BaseURL.JoinPath("DNS", domain, "TXT")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) DeleteTXTRecord(ctx context.Context, domain string, record TXTRecord) error {
	endpoint := c.BaseURL.JoinPath("DNS", domain, "TXT")

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set(authorizationTokenHeader, c.token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
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
