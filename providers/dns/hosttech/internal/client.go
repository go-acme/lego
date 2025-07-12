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
	"golang.org/x/oauth2"
)

const defaultBaseURL = "https://api.ns1.hosttech.eu/api"

// Client a Hosttech client.
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

// GetZones Get a list of all zones.
// https://api.ns1.hosttech.eu/api/documentation/#/Zones/get_api_user_v1_zones
func (c *Client) GetZones(ctx context.Context, query string, limit, offset int) ([]Zone, error) {
	endpoint := c.baseURL.JoinPath("user", "v1", "zones")

	values := endpoint.Query()
	values.Set("query", query)

	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}

	if offset > 0 {
		values.Set("offset", strconv.Itoa(offset))
	}

	endpoint.RawQuery = values.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	result := apiResponse[[]Zone]{}
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetZone Get a single zone.
// https://api.ns1.hosttech.eu/api/documentation/#/Zones/get_api_user_v1_zones__zoneId_
func (c *Client) GetZone(ctx context.Context, zoneID string) (*Zone, error) {
	endpoint := c.baseURL.JoinPath("user", "v1", "zones", zoneID)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	result := apiResponse[*Zone]{}
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetRecords Returns a list of all records for the given zone.
// https://api.ns1.hosttech.eu/api/documentation/#/Records/get_api_user_v1_zones__zoneId__records
func (c *Client) GetRecords(ctx context.Context, zoneID, recordType string) ([]Record, error) {
	endpoint := c.baseURL.JoinPath("user", "v1", "zones", zoneID, "records")

	values := endpoint.Query()

	if recordType != "" {
		values.Set("type", recordType)
	}

	endpoint.RawQuery = values.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	result := apiResponse[[]Record]{}
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// AddRecord Adds a new record to the zone and returns the newly created record.
// https://api.ns1.hosttech.eu/api/documentation/#/Records/post_api_user_v1_zones__zoneId__records
func (c *Client) AddRecord(ctx context.Context, zoneID string, record Record) (*Record, error) {
	endpoint := c.baseURL.JoinPath("user", "v1", "zones", zoneID, "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	result := apiResponse[*Record]{}
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// DeleteRecord Deletes a single record for the given id.
// https://api.ns1.hosttech.eu/api/documentation/#/Records/delete_api_user_v1_zones__zoneId__records__recordId_
func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	endpoint := c.baseURL.JoinPath("user", "v1", "zones", zoneID, "records", recordID)

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
