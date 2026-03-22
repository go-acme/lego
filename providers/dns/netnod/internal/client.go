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
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://primarydnsapi.netnod.se"

// Client the Netnod API client.
type Client struct {
	token string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		token:      token,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) GetZones(ctx context.Context, customerID string, pager Pager) (*PagedResponse[[]Zone], error) {
	endpoint := c.BaseURL.JoinPath("/api/v1/zones")

	values, err := querystring.Values(pager)
	if err != nil {
		return nil, err
	}

	if customerID != "" {
		values.Set("endcustomer", customerID)
	}

	endpoint.RawQuery = values.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(PagedResponse[[]Zone])

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) PartialZoneUpdate(ctx context.Context, zoneID string, rrsets []RRSet) error {
	endpoint := c.BaseURL.JoinPath("/api/v1/zones", zoneID)

	req, err := newJSONRequest(ctx, http.MethodPatch, endpoint, Zone{RRSets: rrsets})
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
	useragent.SetHeader(req.Header)

	req.Header.Set("Authorization", "Token "+c.token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
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
