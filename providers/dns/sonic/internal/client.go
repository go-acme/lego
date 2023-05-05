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
)

const baseURL = "https://public-api.sonic.net/dyndns"

// Client Sonic client.
type Client struct {
	userID string
	apiKey string

	baseURL    string
	HTTPClient *http.Client
}

// NewClient creates a Client.
func NewClient(userID, apiKey string) (*Client, error) {
	if userID == "" || apiKey == "" {
		return nil, errors.New("credentials are missing")
	}

	return &Client{
		userID:     userID,
		apiKey:     apiKey,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// SetRecord creates or updates a TXT records.
// Sonic does not provide a delete record API endpoint.
// https://public-api.sonic.net/dyndns#updating_or_adding_host_records
func (c *Client) SetRecord(ctx context.Context, hostname string, value string, ttl int) error {
	payload := &Record{
		UserID:   c.userID,
		APIKey:   c.apiKey,
		Hostname: hostname,
		Value:    value,
		TTL:      ttl,
		Type:     "TXT",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to create request JSON body: %w", err)
	}

	endpoint, err := url.JoinPath(c.baseURL, "host")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("content-type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	r := APIResponse{}
	err = json.Unmarshal(raw, &r)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if r.Result != 200 {
		return fmt.Errorf("API response code: %d, %s", r.Result, r.Message)
	}

	return nil
}
