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

const defaultBaseURL = "https://api.online.net/api/v1"

// Client the Online API client.
type Client struct {
	apiToken string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiToken string) (*Client, error) {
	if apiToken == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiToken:   apiToken,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateZoneVersion creates a new empty version.
func (c *Client) CreateZoneVersion(ctx context.Context, zone, name string) (*ZoneVersion, error) {
	endpoint := c.BaseURL.JoinPath("domain", zone, "version")

	form := url.Values{}
	form.Set("name", name)

	req, err := newFormRequest(ctx, http.MethodPost, endpoint, form)
	if err != nil {
		return nil, err
	}

	result := new(ZoneVersion)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteZoneVersion deletes a zone version and all associated resource records.
func (c *Client) DeleteZoneVersion(ctx context.Context, zone, versionID string) error {
	endpoint := c.BaseURL.JoinPath("domain", zone, "version", versionID)

	req, err := newFormRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// EnableZoneVersion this will push the version configuration to the DNS server and start propagating the zone.
func (c *Client) EnableZoneVersion(ctx context.Context, zone, versionID string) error {
	endpoint := c.BaseURL.JoinPath("domain", zone, "version", versionID, "enable")

	req, err := newFormRequest(ctx, http.MethodPatch, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// CreateResourceRecord creates a resource record and associates it with a zone.
func (c *Client) CreateResourceRecord(ctx context.Context, zone, versionID string, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("domain", zone, "version", versionID, "zone")

	values, err := querystring.Values(record)
	if err != nil {
		return nil, err
	}

	req, err := newFormRequest(ctx, http.MethodPost, endpoint, values)
	if err != nil {
		return nil, err
	}

	result := new(Record)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteResourceRecord deletes a resource record from a version.
func (c *Client) DeleteResourceRecord(ctx context.Context, zone, versionID, recordID string) error {
	endpoint := c.BaseURL.JoinPath("domain", zone, "version", versionID, "zone", recordID)

	req, err := newFormRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Add("Authorization", "Bearer "+c.apiToken)

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

func newFormRequest(ctx context.Context, method string, endpoint *url.URL, form url.Values) (*http.Request, error) {
	var body io.Reader

	if len(form) > 0 {
		body = bytes.NewReader([]byte(form.Encode()))
	} else {
		body = http.NoBody
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
