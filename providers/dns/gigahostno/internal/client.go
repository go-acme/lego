package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

const defaultBaseURL = "https://api.gigahost.no/api/v0"

// Client is the Gigahost API client.
type Client struct {
	username   string
	password   string
	code       string
	baseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Gigahost API client.
func NewClient(username, password, code string) (*Client, error) {
	if username == "" || password == "" {
		return nil, errors.New("credentials missing")
	}

	return &Client{
		username:   username,
		password:   password,
		code:       code,
		baseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Authenticate obtains an authentication token.
func (c *Client) Authenticate(ctx context.Context) (*Token, error) {
	endpoint := c.baseURL + "/authenticate"

	reqBody := map[string]string{
		"username": c.username,
		"password": c.password,
	}

	if c.code != "" {
		reqBody["code"] = c.code
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse

	err = c.doRequest(req, &authResp)
	if err != nil {
		return nil, err
	}

	if authResp.Data.Token == "" {
		return nil, errors.New("authentication failed: no token received")
	}

	token := &Token{
		Token:    authResp.Data.Token,
		Deadline: time.Unix(authResp.Data.TokenExpire, 0),
	}

	return token, nil
}

// ListZones retrieves all DNS zones.
func (c *Client) ListZones(ctx context.Context, token string) ([]Zone, error) {
	endpoint := c.baseURL + "/dns/zones"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	var zonesResp ZonesResponse

	err = c.doRequest(req, &zonesResp)
	if err != nil {
		return nil, err
	}

	return zonesResp.Data, nil
}

// ListRecords retrieves all records for a zone.
func (c *Client) ListRecords(ctx context.Context, token, zoneID string) ([]Record, error) {
	endpoint := fmt.Sprintf("%s/dns/zones/%s/records", c.baseURL, zoneID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	var recordsResp RecordsResponse

	err = c.doRequest(req, &recordsResp)
	if err != nil {
		return nil, err
	}

	return recordsResp.Data, nil
}

// CreateRecord creates a new DNS record.
func (c *Client) CreateRecord(ctx context.Context, token, zoneID string, record CreateRecordRequest) error {
	endpoint := fmt.Sprintf("%s/dns/zones/%s/records", c.baseURL, zoneID)

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	return c.doRequest(req, nil)
}

// DeleteRecord deletes a DNS record.
func (c *Client) DeleteRecord(ctx context.Context, token, zoneID, recordID, name, recordType string) error {
	endpoint := fmt.Sprintf("%s/dns/zones/%s/records/%s?name=%s&type=%s",
		c.baseURL, zoneID, recordID, name, recordType)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	return c.doRequest(req, nil)
}

func (c *Client) doRequest(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return parseError(req, resp)
	}

	if result == nil {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(body, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, body, err)
	}

	return nil
}

// newJSONRequest creates a new JSON HTTP request.
func newJSONRequest(ctx context.Context, method, url string, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// parseError extracts error information from response.
func parseError(req *http.Request, resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, body)
}
