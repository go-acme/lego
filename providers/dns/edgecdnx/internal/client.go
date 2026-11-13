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

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

const defaultBaseURL = "https://api.edgecdnx.com"

// Client the EdgeCDN-X API client.
type Client struct {
	apiKey    string
	projectID string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(baseURL, apiKey, projectID string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	if projectID == "" {
		return nil, errors.New("missing project ID")
	}

	bu, err := getBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		apiKey:     apiKey,
		projectID:  projectID,
		baseURL:    bu,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// ListZones lists zones.
// https://api.edgecdnx.com/swagger/index.html#/Zones/get_project__project_id__zones
func (c *Client) ListZones(ctx context.Context) ([]Zone, error) {
	endpoint := c.baseURL.JoinPath("project", c.projectID, "zones")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var zones []Zone

	err = c.do(req, &zones)
	if err != nil {
		return nil, err
	}

	return zones, nil
}

// ListDNSEndpoints lists DNS endpoints for a zone.
// https://api.edgecdnx.com/swagger/index.html#/DNS%20Endpoints/get_project__project_id__zones__zone_id__dns_endpoints
func (c *Client) ListDNSEndpoints(ctx context.Context, zoneID string) ([]Record, error) {
	endpoint := c.baseURL.JoinPath("project", c.projectID, "zones", zoneID, "dns-endpoints")

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

// CreateDNSEndpoint creates a DNS endpoint.
// https://api.edgecdnx.com/swagger/index.html#/DNS%20Endpoints/post_project__project_id__zones__zone_id__dns_endpoints
func (c *Client) CreateDNSEndpoint(ctx context.Context, zoneID string, record RecordSpec) error {
	endpoint := c.baseURL.JoinPath("project", c.projectID, "zones", zoneID, "dns-endpoints")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// UpdateDNSEndpoint updates a DNS endpoint.
// https://api.edgecdnx.com/swagger/index.html#/DNS%20Endpoints/patch_project__project_id__zones__zone_id__dns_endpoints__dns_endpoint_id_
func (c *Client) UpdateDNSEndpoint(ctx context.Context, zoneID, endpointID string, record RecordSpec) error {
	endpoint := c.baseURL.JoinPath("project", c.projectID, "zones", zoneID, "dns-endpoints", endpointID)

	req, err := newJSONRequest(ctx, http.MethodPatch, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// DeleteDNSEndpoint deletes a DNS endpoint.
// https://api.edgecdnx.com/swagger/index.html#/DNS%20Endpoints/delete_project__project_id__zones__zone_id__dns_endpoints__dns_endpoint_id_
func (c *Client) DeleteDNSEndpoint(ctx context.Context, zoneID, endpointID, resourceVersion string) error {
	endpoint := c.baseURL.JoinPath("project", c.projectID, "zones", zoneID, "dns-endpoints", endpointID)

	if resourceVersion != "" {
		query := endpoint.Query()
		query.Set("resourceVersion", resourceVersion)
		endpoint.RawQuery = query.Encode()
	}

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set("Authorization", "ApiKey "+c.apiKey)

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

	errAPI := &APIError{StatusCode: resp.StatusCode}

	err := json.Unmarshal(raw, errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return errAPI
}

func getBaseURL(baseURL string) (*url.URL, error) {
	if baseURL == "" {
		return url.Parse(defaultBaseURL)
	}

	return url.Parse(baseURL)
}
