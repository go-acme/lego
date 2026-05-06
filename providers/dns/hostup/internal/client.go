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

	"github.com/go-acme/lego/v5/internal/errutils"
	"golang.org/x/oauth2"
)

const defaultBaseURL = "https://cloud.hostup.se/api"

// Client a HostUp client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new Client.
func NewClient(hc *http.Client) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{baseURL: baseURL, httpClient: hc}
}

// GetZones returns the zones available to the API key.
// https://hostup.se/en/support/api-autentisering/
func (c *Client) GetZones(ctx context.Context) ([]Zone, error) {
	endpoint := c.baseURL.JoinPath("dns", "zones")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	result := apiResponse[zonesData]{}

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data.Zones, nil
}

// AddRecord creates a new record in the given zone.
func (c *Client) AddRecord(ctx context.Context, zoneID string, record Record) (*Record, error) {
	endpoint := c.baseURL.JoinPath("dns", "zones", zoneID, "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	result := apiResponse[recordData]{}

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data.Record, nil
}

// DeleteRecord deletes a record by ID in the given zone.
func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	endpoint := c.baseURL.JoinPath("dns", "zones", zoneID, "records", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	resp, errD := c.httpClient.Do(req)
	if errD != nil {
		return errutils.NewHTTPDoError(req, errD)
	}

	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
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

	case http.StatusNoContent:
		return nil

	default:
		return parseError(req, resp)
	}
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

	errAPI := &APIError{StatusCode: resp.StatusCode}

	err := json.Unmarshal(raw, errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return errAPI
}

// OAuthStaticAccessToken wraps an HTTP client with a static bearer token.
func OAuthStaticAccessToken(client *http.Client, accessToken string) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	client.Transport = &oauth2.Transport{
		Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
		Base:   client.Transport,
	}

	return client
}
