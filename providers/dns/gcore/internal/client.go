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
)

const defaultBaseURL = "https://api.gcore.com/dns"

const (
	authorizationHeader = "Authorization"
	tokenTypeHeader     = "APIKey"
)

const txtRecordType = "TXT"

// Client for DNS API.
type Client struct {
	token string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient constructor of Client.
func NewClient(token string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		token:      token,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetZone gets zone information.
// https://api.gcore.com/docs/dns#tag/zones/operation/Zone
func (c *Client) GetZone(ctx context.Context, name string) (Zone, error) {
	endpoint := c.baseURL.JoinPath("v2", "zones", name)

	zone := Zone{}
	err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &zone)
	if err != nil {
		return Zone{}, fmt.Errorf("get zone %s: %w", name, err)
	}

	return zone, nil
}

// GetRRSet gets RRSet item.
// https://api.gcore.com/docs/dns#tag/rrsets/operation/RRSet
func (c *Client) GetRRSet(ctx context.Context, zone, name string) (RRSet, error) {
	endpoint := c.baseURL.JoinPath("v2", "zones", zone, name, txtRecordType)

	var result RRSet
	err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &result)
	if err != nil {
		return RRSet{}, fmt.Errorf("get txt records %s -> %s: %w", zone, name, err)
	}

	return result, nil
}

// DeleteRRSet removes RRSet record.
// https://api.gcore.com/docs/dns#tag/rrsets/operation/DeleteRRSet
func (c *Client) DeleteRRSet(ctx context.Context, zone, name string) error {
	endpoint := c.baseURL.JoinPath("v2", "zones", zone, name, txtRecordType)

	err := c.doRequest(ctx, http.MethodDelete, endpoint, nil, nil)
	if err != nil {
		// Support DELETE idempotence https://developer.mozilla.org/en-US/docs/Glossary/Idempotent
		statusErr := new(APIError)
		if errors.As(err, statusErr) && statusErr.StatusCode == http.StatusNotFound {
			return nil
		}

		return fmt.Errorf("delete record request: %w", err)
	}

	return nil
}

// AddRRSet adds TXT record (create or update).
func (c *Client) AddRRSet(ctx context.Context, zone, recordName, value string, ttl int) error {
	record := RRSet{TTL: ttl, Records: []Records{{Content: []string{value}}}}

	txt, err := c.GetRRSet(ctx, zone, recordName)
	if err == nil && len(txt.Records) > 0 {
		record.Records = append(record.Records, txt.Records...)
		return c.updateRRSet(ctx, zone, recordName, record)
	}

	return c.createRRSet(ctx, zone, recordName, record)
}

// https://api.gcore.com/docs/dns#tag/rrsets/operation/CreateRRSet
func (c *Client) createRRSet(ctx context.Context, zone, name string, record RRSet) error {
	endpoint := c.baseURL.JoinPath("v2", "zones", zone, name, txtRecordType)

	return c.doRequest(ctx, http.MethodPost, endpoint, record, nil)
}

// https://api.gcore.com/docs/dns#tag/rrsets/operation/UpdateRRSet
func (c *Client) updateRRSet(ctx context.Context, zone, name string, record RRSet) error {
	endpoint := c.baseURL.JoinPath("v2", "zones", zone, name, txtRecordType)

	return c.doRequest(ctx, http.MethodPut, endpoint, record, nil)
}

func (c *Client) doRequest(ctx context.Context, method string, endpoint *url.URL, bodyParams any, result any) error {
	req, err := newJSONRequest(ctx, method, endpoint, bodyParams)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set(authorizationHeader, fmt.Sprintf("%s %s", tokenTypeHeader, c.token))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return parseError(resp)
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

func parseError(resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	errAPI := APIError{StatusCode: resp.StatusCode}
	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		errAPI.Message = string(raw)
	}

	return errAPI
}
