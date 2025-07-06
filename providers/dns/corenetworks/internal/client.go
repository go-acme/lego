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

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://beta.api.core-networks.de"

// Client a Core-Networks client.
type Client struct {
	login    string
	password string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(login, password string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		login:      login,
		password:   password,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// ListZone gets a list of all DNS zones.
// https://beta.api.core-networks.de/doc/#functon_dnszones
func (c *Client) ListZone(ctx context.Context) ([]Zone, error) {
	endpoint := c.baseURL.JoinPath("dnszones")

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

// GetZoneDetails provides detailed information about a DNS zone.
// https://beta.api.core-networks.de/doc/#functon_dnszones_details
func (c *Client) GetZoneDetails(ctx context.Context, zone string) (*ZoneDetails, error) {
	endpoint := c.baseURL.JoinPath("dnszones", zone)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var details ZoneDetails
	err = c.do(req, &details)
	if err != nil {
		return nil, err
	}

	return &details, nil
}

// ListRecords gets a list of DNS records belonging to the zone.
// https://beta.api.core-networks.de/doc/#functon_dnszones_records
func (c *Client) ListRecords(ctx context.Context, zone string) ([]Record, error) {
	endpoint := c.baseURL.JoinPath("dnszones", zone, "records")

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

// AddRecord adds a record.
// https://beta.api.core-networks.de/doc/#functon_dnszones_records_add
func (c *Client) AddRecord(ctx context.Context, zone string, record Record) error {
	endpoint := c.baseURL.JoinPath("dnszones", zone, "records", "/")

	if record.Name == "" {
		record.Name = "@"
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

// DeleteRecords deletes all DNS records of a zone that match the DNS record passed.
// https://beta.api.core-networks.de/doc/#functon_dnszones_records_delete
func (c *Client) DeleteRecords(ctx context.Context, zone string, record Record) error {
	endpoint := c.baseURL.JoinPath("dnszones", zone, "records", "delete")

	if record.Name == "" {
		record.Name = "@"
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

// CommitRecords sends a commit to the zone.
// https://beta.api.core-networks.de/doc/#functon_dnszones_commit
func (c *Client) CommitRecords(ctx context.Context, zone string) error {
	endpoint := c.baseURL.JoinPath("dnszones", zone, "records", "commit")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	at := getToken(req.Context())
	if at != "" {
		req.Header.Set(authorizationHeader, "Bearer "+at)
	}

	resp, errD := c.HTTPClient.Do(req)
	if errD != nil {
		return errutils.NewHTTPDoError(req, errD)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
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
