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

	"github.com/cenkalti/backoff/v4"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://secure.hosting.de/api/dns/v1/json"

// Client the API client for Hosting.de.
type Client struct {
	apiKey string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates new Client.
func NewClient(apiKey string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetZone gets a zone.
func (c Client) GetZone(ctx context.Context, req ZoneConfigsFindRequest) (*ZoneConfig, error) {
	var zoneConfig *ZoneConfig

	operation := func() error {
		response, err := c.ListZoneConfigs(ctx, req)
		if err != nil {
			return backoff.Permanent(err)
		}

		if response.Data[0].Status != "active" {
			return fmt.Errorf("unexpected status: %q", response.Data[0].Status)
		}

		zoneConfig = &response.Data[0]

		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 3 * time.Second
	bo.MaxInterval = 10 * bo.InitialInterval
	bo.MaxElapsedTime = 100 * bo.InitialInterval

	// retry in case the zone was edited recently and is not yet active
	err := backoff.Retry(operation, bo)
	if err != nil {
		return nil, err
	}

	return zoneConfig, nil
}

// ListZoneConfigs lists zone configuration.
// https://www.hosting.de/api/?json#list-zoneconfigs
func (c Client) ListZoneConfigs(ctx context.Context, req ZoneConfigsFindRequest) (*ZoneResponse, error) {
	endpoint := c.baseURL.JoinPath("zoneConfigsFind")

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
func (c Client) UpdateZone(ctx context.Context, req ZoneUpdateRequest) (*Zone, error) {
	endpoint := c.baseURL.JoinPath("zoneUpdate")

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

func (c Client) post(ctx context.Context, endpoint *url.URL, request, result any) ([]byte, error) {
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
