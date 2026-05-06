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
	"strconv"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

const defaultBaseURL = "https://api.ngenix.net/api/v3"

// Client the Ngenix API client.
type Client struct {
	username   string
	token      string
	customerID string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(username, token, customerID string) (*Client, error) {
	if username == "" {
		return nil, errors.New("credentials missing: username")
	}

	if token == "" {
		return nil, errors.New("credentials missing: token")
	}

	if customerID == "" {
		return nil, errors.New("credentials missing: customerID")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		username:   username,
		token:      token,
		customerID: customerID,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// ListDNSZones lists all DNS zones for the customer.
func (c *Client) ListDNSZones(ctx context.Context) ([]DNSZoneCollectionView, error) {
	endpoint := c.BaseURL.JoinPath("dns-zone")

	query := endpoint.Query()
	query.Set("customerId", c.customerID)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var collection DNSZoneCollection

	if err := c.do(req, &collection); err != nil {
		return nil, err
	}

	return collection.Elements, nil
}

// GetDNSZone gets a DNS zone by ID including all its records.
func (c *Client) GetDNSZone(ctx context.Context, zoneID int64) (*DNSZone, error) {
	endpoint := c.BaseURL.JoinPath("dns-zone", strconv.FormatInt(zoneID, 10))

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	zone := new(DNSZone)

	if err := c.do(req, zone); err != nil {
		return nil, err
	}

	return zone, nil
}

// UpdateDNSZone replaces all records of a DNS zone.
func (c *Client) UpdateDNSZone(ctx context.Context, zoneID int64, update DNSZoneUpdate) (*DNSZone, error) {
	endpoint := c.BaseURL.JoinPath("dns-zone", strconv.FormatInt(zoneID, 10))

	req, err := newJSONRequest(ctx, http.MethodPatch, endpoint, update)
	if err != nil {
		return nil, err
	}

	zone := new(DNSZone)

	if err := c.do(req, zone); err != nil {
		return nil, err
	}

	return zone, nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.SetBasicAuth(c.username+"/token", c.token)

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

	errAPI := new(APIError)

	err := json.Unmarshal(raw, errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	if errAPI.Code == 0 {
		errAPI.Code = resp.StatusCode
	}

	return errAPI
}
