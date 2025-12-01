package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// DefaultEndpoint default API endpoint.
const DefaultEndpoint = "https://api.autodns.com/v1/"

// DefaultEndpointContext default API endpoint context.
const DefaultEndpointContext int = 4

// Client the Autodns API client.
type Client struct {
	username string
	password string
	context  int

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(username, password string, clientContext int) *Client {
	baseURL, _ := url.Parse(DefaultEndpoint)

	return &Client{
		username:   username,
		password:   password,
		context:    clientContext,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// AddRecords adds records.
func (c *Client) AddRecords(ctx context.Context, domain string, records []*ResourceRecord) (*DataZoneResponse, error) {
	zoneStream := &ZoneStream{Adds: records}

	return c.updateZone(ctx, domain, zoneStream)
}

// RemoveRecords removes records.
func (c *Client) RemoveRecords(ctx context.Context, domain string, records []*ResourceRecord) (*DataZoneResponse, error) {
	zoneStream := &ZoneStream{Removes: records}

	return c.updateZone(ctx, domain, zoneStream)
}

// https://github.com/InterNetX/domainrobot-api/blob/bdc8fe92a2f32fcbdb29e30bf6006ab446f81223/src/domainrobot.json#L21090
func (c *Client) updateZone(ctx context.Context, domain string, zoneStream *ZoneStream) (*DataZoneResponse, error) {
	endpoint := c.BaseURL.JoinPath("zone", domain, "_stream")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, zoneStream)
	if err != nil {
		return nil, err
	}

	var resp *DataZoneResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set("X-Domainrobot-Context", strconv.Itoa(c.context))
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return parseError(req, resp)
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

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	var errAPI APIError

	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return &errAPI
}
