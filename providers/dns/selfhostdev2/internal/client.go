package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
)

const defaultBaseURL = "https://my.selfhost.de/cgi-bin/dns-api.pl"

// Client the SelfHost client.
type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient() *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// UpdateTXTRecord updates content of an existing TXT record.
func (c *Client) UpdateTXTRecord(ctx context.Context, payload Payload) error {
	req, err := newJSONRequest(ctx, c.BaseURL, payload)
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

func newJSONRequest(ctx context.Context, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
