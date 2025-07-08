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
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://api.dynect.net/REST"

// Client the Dyn API client.
type Client struct {
	customerName string
	username     string
	password     string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(customerName, username, password string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		customerName: customerName,
		username:     username,
		password:     password,
		baseURL:      baseURL,
		HTTPClient:   &http.Client{Timeout: 5 * time.Second},
	}
}

// Publish updating Zone settings.
// https://help.dyn.com/update-zone-api/
func (c *Client) Publish(ctx context.Context, zone, notes string) error {
	endpoint := c.baseURL.JoinPath("Zone", zone)

	payload := &publish{Publish: true, Notes: notes}

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, payload)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

// AddTXTRecord creating TXT Records.
// https://help.dyn.com/create-txt-record-api/
func (c *Client) AddTXTRecord(ctx context.Context, authZone, fqdn, value string, ttl int) error {
	endpoint := c.baseURL.JoinPath("TXTRecord", authZone, fqdn)

	payload := map[string]any{
		"rdata": map[string]string{
			"txtdata": value,
		},
		"ttl": strconv.Itoa(ttl),
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

// RemoveTXTRecord deleting one or all existing TXT Records.
// https://help.dyn.com/delete-txt-records-api/
func (c *Client) RemoveTXTRecord(ctx context.Context, authZone, fqdn string) error {
	endpoint := c.baseURL.JoinPath("TXTRecord", authZone, fqdn)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	return nil
}

func (c *Client) do(req *http.Request) (*APIResponse, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var response APIResponse
	err = json.Unmarshal(raw, &response)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("%s: %w", response.Messages, errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw))
	}

	if resp.StatusCode == http.StatusTemporaryRedirect {
		// TODO add support for HTTP 307 response and long running jobs
		return nil, errors.New("API request returned HTTP 307. This is currently unsupported")
	}

	if response.Status == "failure" {
		return nil, fmt.Errorf("%s: %w", response.Messages, errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw))
	}

	return &response, nil
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

	tok := getToken(req.Context())
	if tok != "" {
		req.Header.Set(authTokenHeader, tok)
	}

	return req, nil
}
