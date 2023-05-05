package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"golang.org/x/net/publicsuffix"
)

const defaultBaseURL = "https://gateway.stackpath.com/dns/v1/stacks/"

// Client the API client for Stackpath.
type Client struct {
	stackID string

	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new Client.
func NewClient(ctx context.Context, stackID, clientID, clientSecret string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		baseURL:    baseURL,
		stackID:    stackID,
		httpClient: createOAuthClient(ctx, clientID, clientSecret),
	}
}

// GetZones gets all zones.
// https://stackpath.dev/reference/getzones
func (c *Client) GetZones(ctx context.Context, domain string) (*Zone, error) {
	endpoint := c.baseURL.JoinPath(c.stackID, "zones")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	tld, err := publicsuffix.EffectiveTLDPlusOne(dns01.UnFqdn(domain))
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("page_request.filter", fmt.Sprintf("domain='%s'", tld))
	req.URL.RawQuery = query.Encode()

	var zones Zones
	err = c.do(req, &zones)
	if err != nil {
		return nil, err
	}

	if len(zones.Zones) == 0 {
		return nil, fmt.Errorf("did not find zone with domain %s", domain)
	}

	return &zones.Zones[0], nil
}

// GetZoneRecords gets all records.
// https://stackpath.dev/reference/getzonerecords
func (c *Client) GetZoneRecords(ctx context.Context, name string, zone *Zone) ([]Record, error) {
	endpoint := c.baseURL.JoinPath(c.stackID, "zones", zone.ID, "records")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("page_request.filter", fmt.Sprintf("name='%s' and type='TXT'", name))
	req.URL.RawQuery = query.Encode()

	var records Records
	err = c.do(req, &records)
	if err != nil {
		return nil, err
	}

	if len(records.Records) == 0 {
		return nil, fmt.Errorf("did not find record with name %s", name)
	}

	return records.Records, nil
}

// CreateZoneRecord creates a record.
// https://stackpath.dev/reference/createzonerecord
func (c *Client) CreateZoneRecord(ctx context.Context, zone *Zone, record Record) error {
	endpoint := c.baseURL.JoinPath(c.stackID, "zones", zone.ID, "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// DeleteZoneRecord deletes a record.
// https://stackpath.dev/reference/deletezonerecord
func (c *Client) DeleteZoneRecord(ctx context.Context, zone *Zone, record Record) error {
	endpoint := c.baseURL.JoinPath(c.stackID, "zones", zone.ID, "records", record.ID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
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

func (c *Client) do(req *http.Request, result any) error {
	resp, err := c.httpClient.Do(req)
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

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	errResp := &ErrorResponse{}
	err := json.Unmarshal(raw, errResp)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return errResp
}
