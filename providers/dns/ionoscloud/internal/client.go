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
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

const defaultBaseURL = "https://dns.de-fra.ionos.com"

const authorizationHeader = "Authorization"

// Client the Ionos Cloud API client.
type Client struct {
	apiKey string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// RetrieveZones returns a list of the DNS zones.
// https://api.ionos.com/docs/dns/v1/#tag/Zones/operation/zonesGet
func (c *Client) RetrieveZones(ctx context.Context, zoneName string) ([]Zone, error) {
	endpoint := c.BaseURL.JoinPath("zones")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("filter.zoneName", zoneName)
	req.URL.RawQuery = query.Encode()

	result := ZonesResponse{}

	if err := c.do(req, &result); err != nil {
		return nil, err
	}

	return result.Items, nil
}

// CreateRecord creates a new record for the DNS zone.
// https://api.ionos.com/docs/dns/v1/#tag/Records/operation/zonesRecordsPost
func (c *Client) CreateRecord(ctx context.Context, zoneID string, record RecordProperties) (*RecordResponse, error) {
	endpoint := c.BaseURL.JoinPath("zones", zoneID, "records")

	payload := map[string]RecordProperties{
		"properties": record,
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}

	result := &RecordResponse{}

	if err := c.do(req, result); err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteRecord deletes a specified record from the DNS zone.
// https://api.ionos.com/docs/dns/v1/#tag/Records/operation/zonesRecordsDelete
func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	endpoint := c.BaseURL.JoinPath("zones", zoneID, "records", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set(authorizationHeader, "Bearer "+c.apiKey)

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
