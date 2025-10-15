package hostingde

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const (
	DefaultHostingdeBaseURL = "https://secure.hosting.de/api/dns/v1/json"
	DefaultHTTPNetBaseURL   = "https://partner.http.net/api/dns/v1/json"
)

// Client the API client for Hosting.de.
type Client struct {
	apiKey string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates new Client.
func NewClient(apiKey string) *Client {
	baseURL, _ := url.Parse(DefaultHostingdeBaseURL)

	return &Client{
		apiKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetZone gets a zone.
func (c *Client) GetZone(ctx context.Context, req ZoneConfigsFindRequest) (*ZoneConfig, error) {
	operation := func() (*ZoneConfig, error) {
		response, err := c.ListZoneConfigs(ctx, req)
		if err != nil {
			return nil, backoff.Permanent(err)
		}

		if response.Data[0].Status != "active" {
			return nil, fmt.Errorf("unexpected status: %q", response.Data[0].Status)
		}

		return &response.Data[0], nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 3 * time.Second
	bo.MaxInterval = 10 * bo.InitialInterval

	// retry in case the zone was edited recently and is not yet active
	return backoff.Retry(ctx, operation, backoff.WithBackOff(bo), backoff.WithMaxElapsedTime(100*bo.InitialInterval))
}

// ListZoneConfigs lists zone configuration.
// https://www.hosting.de/api/?json#list-zoneconfigs
func (c *Client) ListZoneConfigs(ctx context.Context, req ZoneConfigsFindRequest) (*ZoneResponse, error) {
	endpoint := c.BaseURL.JoinPath("zoneConfigsFind")

	req.AuthToken = c.apiKey

	response := &BaseResponse[*ZoneResponse]{}

	rawResp, err := c.post(ctx, endpoint, req, response)
	if err != nil {
		return nil, err
	}

	if response.Status != "success" && response.Status != "pending" {
		return nil, fmt.Errorf("unexpected status: %q, %s", response.Status, string(rawResp))
	}

	if response.Response == nil || len(response.Response.Data) == 0 {
		return nil, fmt.Errorf("no data, status: %q, %s", response.Status, string(rawResp))
	}

	return response.Response, nil
}

// UpdateZone updates a zone.
// https://www.hosting.de/api/?json#updating-zones
func (c *Client) UpdateZone(ctx context.Context, req ZoneUpdateRequest) (*Zone, error) {
	endpoint := c.BaseURL.JoinPath("zoneUpdate")

	req.AuthToken = c.apiKey

	// but we'll need the ID later to delete the record
	response := &BaseResponse[*Zone]{}

	rawResp, err := c.post(ctx, endpoint, req, response)
	if err != nil {
		return nil, err
	}

	if response.Status != "success" && response.Status != "pending" {
		return nil, fmt.Errorf("unexpected status: %q, %s", response.Status, string(rawResp))
	}

	return response.Response, nil
}

func (c *Client) post(ctx context.Context, endpoint *url.URL, request, result any) ([]byte, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to create request JSON body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return raw, nil
}
