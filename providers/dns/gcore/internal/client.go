package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.gcorelabs.com/dns"
	tokenHeader    = "APIKey"
	txtRecordType  = "TXT"
)

// Client for DNS API.
type Client struct {
	HTTPClient *http.Client
	baseURL    *url.URL
	token      string
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
// https://dnsapi.gcorelabs.com/docs#operation/Zone
func (c *Client) GetZone(ctx context.Context, name string) (Zone, error) {
	zone := Zone{}
	uri := path.Join("/v2/zones", name)

	err := c.do(ctx, http.MethodGet, uri, nil, &zone)
	if err != nil {
		return Zone{}, fmt.Errorf("get zone %s: %w", name, err)
	}

	return zone, nil
}

// GetRRSet gets RRSet item.
// https://dnsapi.gcorelabs.com/docs#operation/RRSet
func (c *Client) GetRRSet(ctx context.Context, zone, name string) (RRSet, error) {
	var result RRSet
	uri := path.Join("/v2/zones", zone, name, txtRecordType)

	err := c.do(ctx, http.MethodGet, uri, nil, &result)
	if err != nil {
		return RRSet{}, fmt.Errorf("get txt records %s -> %s: %w", zone, name, err)
	}

	return result, nil
}

// DeleteRRSet removes RRSet record.
// https://dnsapi.gcorelabs.com/docs#operation/DeleteRRSet
func (c *Client) DeleteRRSet(ctx context.Context, zone, name string) error {
	uri := path.Join("/v2/zones", zone, name, txtRecordType)

	err := c.do(ctx, http.MethodDelete, uri, nil, nil)
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

// https://dnsapi.gcorelabs.com/docs#operation/CreateRRSet
func (c *Client) createRRSet(ctx context.Context, zone, name string, record RRSet) error {
	uri := path.Join("/v2/zones", zone, name, txtRecordType)

	return c.do(ctx, http.MethodPost, uri, record, nil)
}

// https://dnsapi.gcorelabs.com/docs#operation/UpdateRRSet
func (c *Client) updateRRSet(ctx context.Context, zone, name string, record RRSet) error {
	uri := path.Join("/v2/zones", zone, name, txtRecordType)

	return c.do(ctx, http.MethodPut, uri, record, nil)
}

func (c *Client) do(ctx context.Context, method, uri string, bodyParams interface{}, dest interface{}) error {
	var bs []byte
	if bodyParams != nil {
		var err error
		bs, err = json.Marshal(bodyParams)
		if err != nil {
			return fmt.Errorf("encode bodyParams: %w", err)
		}
	}

	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, uri))
	if err != nil {
		return fmt.Errorf("failed to parse endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), strings.NewReader(string(bs)))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", tokenHeader, c.token))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		all, _ := io.ReadAll(resp.Body)

		e := APIError{
			StatusCode: resp.StatusCode,
		}

		err := json.Unmarshal(all, &e)
		if err != nil {
			e.Message = string(all)
		}

		return e
	}

	if dest == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}
