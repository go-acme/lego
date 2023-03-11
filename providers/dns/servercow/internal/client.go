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

const baseAPIURL = "https://api.servercow.de/dns/v1/domains"

// Client the Servercow client.
type Client struct {
	username string
	password string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a Servercow client.
func NewClient(username, password string) *Client {
	baseURL, _ := url.Parse(baseAPIURL)

	return &Client{
		username:   username,
		password:   password,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetRecords from API.
func (c *Client) GetRecords(ctx context.Context, domain string) ([]Record, error) {
	endpoint := c.baseURL.JoinPath(domain)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var records []Record
	err = c.do(req, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// CreateUpdateRecord creates or updates a record.
func (c *Client) CreateUpdateRecord(ctx context.Context, domain string, data Record) (*Message, error) {
	endpoint := c.baseURL.JoinPath(domain)

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, data)
	if err != nil {
		return nil, err
	}

	var msg Message
	err = c.do(req, &msg)
	if err != nil {
		return nil, err
	}

	if msg.ErrorMsg != "" {
		return nil, msg
	}

	return &msg, nil
}

// DeleteRecord deletes a record.
func (c *Client) DeleteRecord(ctx context.Context, domain string, data Record) (*Message, error) {
	endpoint := c.baseURL.JoinPath(domain)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, data)
	if err != nil {
		return nil, err
	}

	var msg Message
	err = c.do(req, &msg)
	if err != nil {
		return nil, err
	}

	if msg.ErrorMsg != "" {
		return nil, msg
	}

	return &msg, nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set("X-Auth-Username", c.username)
	req.Header.Set("X-Auth-Password", c.password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	// Note the API always return 200 even if the authentication failed.
	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = unmarshal(raw, result)
	if err != nil {
		return err
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

func unmarshal(raw []byte, v any) error {
	err := json.Unmarshal(raw, v)
	if err == nil {
		return nil
	}

	var utErr *json.UnmarshalTypeError

	if !errors.As(err, &utErr) {
		return fmt.Errorf("unmarshaling %T error: %w: %s", v, err, string(raw))
	}

	var apiErr Message
	errU := json.Unmarshal(raw, &apiErr)
	if errU != nil {
		return fmt.Errorf("unmarshaling %T error: %w: %s", v, err, string(raw))
	}

	return apiErr
}
