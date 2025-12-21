package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

type Client struct {
	serverURL  string
	HTTPClient *http.Client
}

func NewClient(serverURL string) (*Client, error) {
	_, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("server URL: %w", err)
	}

	return &Client{
		serverURL:  serverURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) Login(ctx context.Context, username, password string) (string, error) {
	payload := LoginRequest{
		Username:    username,
		Password:    password,
		ClientLogin: false,
	}

	endpoint, err := url.Parse(c.serverURL)
	if err != nil {
		return "", err
	}

	endpoint.RawQuery = "login"

	req, err := newJSONRequest(ctx, endpoint, payload)
	if err != nil {
		return "", err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return "", err
	}

	return extractResponse[string](response)
}

func (c *Client) GetClientID(ctx context.Context, sessionID, sysUserID string) (int, error) {
	payload := ClientIDRequest{
		SessionID: sessionID,
		SysUserID: sysUserID,
	}

	endpoint, err := url.Parse(c.serverURL)
	if err != nil {
		return 0, err
	}

	endpoint.RawQuery = "client_get_id"

	req, err := newJSONRequest(ctx, endpoint, payload)
	if err != nil {
		return 0, err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return 0, err
	}

	return extractResponse[int](response)
}

// GetZoneID returns the zone ID for the given name.
func (c *Client) GetZoneID(ctx context.Context, sessionID, name string) (int, error) {
	payload := map[string]any{
		"session_id": sessionID,
		"origin":     name,
	}

	endpoint, err := url.Parse(c.serverURL)
	if err != nil {
		return 0, err
	}

	endpoint.RawQuery = "dns_zone_get_id"

	req, err := newJSONRequest(ctx, endpoint, payload)
	if err != nil {
		return 0, err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return 0, err
	}

	return extractResponse[int](response)
}

// GetZone returns the zone information for the zone ID.
func (c *Client) GetZone(ctx context.Context, sessionID, zoneID string) (*Zone, error) {
	payload := map[string]any{
		"session_id": sessionID,
		"primary_id": zoneID,
	}

	endpoint, err := url.Parse(c.serverURL)
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = "dns_zone_get"

	req, err := newJSONRequest(ctx, endpoint, payload)
	if err != nil {
		return nil, err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return nil, err
	}

	return extractResponse[*Zone](response)
}

// GetTXT returns the TXT record for the given name.
// `name` must be a fully qualified domain name, e.g. "example.com.".
func (c *Client) GetTXT(ctx context.Context, sessionID, name string) (*Record, error) {
	payload := GetTXTRequest{
		SessionID: sessionID,
		PrimaryID: struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}{
			Name: name,
			Type: "txt",
		},
	}

	endpoint, err := url.Parse(c.serverURL)
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = "dns_txt_get"

	req, err := newJSONRequest(ctx, endpoint, payload)
	if err != nil {
		return nil, err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return nil, err
	}

	return extractResponse[*Record](response)
}

// AddTXT adds a TXT record.
// It returns the ID of the newly created record.
func (c *Client) AddTXT(ctx context.Context, sessionID, clientID string, params RecordParams) (string, error) {
	payload := AddTXTRequest{
		SessionID:    sessionID,
		ClientID:     clientID,
		Params:       &params,
		UpdateSerial: true,
	}

	endpoint, err := url.Parse(c.serverURL)
	if err != nil {
		return "", err
	}

	endpoint.RawQuery = "dns_txt_add"

	req, err := newJSONRequest(ctx, endpoint, payload)
	if err != nil {
		return "", err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return "", err
	}

	return extractResponse[string](response)
}

// DeleteTXT deletes a TXT record.
// It returns the number of deleted records.
func (c *Client) DeleteTXT(ctx context.Context, sessionID, recordID string) (int, error) {
	payload := DeleteTXTRequest{
		SessionID:    sessionID,
		PrimaryID:    recordID,
		UpdateSerial: true,
	}

	endpoint, err := url.Parse(c.serverURL)
	if err != nil {
		return 0, err
	}

	endpoint.RawQuery = "dns_txt_delete"

	req, err := newJSONRequest(ctx, endpoint, payload)
	if err != nil {
		return 0, err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return 0, err
	}

	return extractResponse[int](response)
}

func (c *Client) do(req *http.Request, result any) error {
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

func extractResponse[T any](response APIResponse) (T, error) {
	if response.Code != "ok" {
		var zero T

		return zero, &APIError{APIResponse: response}
	}

	var result T

	err := json.Unmarshal(response.Response, &result)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("unable to unmarshal response: %s, %w", string(response.Response), err)
	}

	return result, nil
}
