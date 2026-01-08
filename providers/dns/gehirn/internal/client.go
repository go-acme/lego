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

const defaultBaseURL = "https://api.gis.gehirn.jp/dns/v1"

// Client the Gehirn API client.
type Client struct {
	tokenID     string
	tokenSecret string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(tokenID, tokenSecret string) (*Client, error) {
	if tokenID == "" || tokenSecret == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		tokenID:     tokenID,
		tokenSecret: tokenSecret,
		BaseURL:     baseURL,
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateRecord creates a record.
// https://support.gehirn.jp/apidocs/dns/records.html#post-dns-v1-records
func (c *Client) CreateRecord(ctx context.Context, zoneID, versionID string, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("zones", zoneID, "versions", versionID, "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := &Record{}

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteRecord deletes a record.
// https://support.gehirn.jp/apidocs/dns/records.html#delete-dns-v1-records
func (c *Client) DeleteRecord(ctx context.Context, zoneID, versionID, recordID string) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("zones", zoneID, "versions", versionID, "records", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := &Record{}

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListZones lists all zones.
// https://support.gehirn.jp/apidocs/dns/zones.html#list
func (c *Client) ListZones(ctx context.Context) ([]Zone, error) {
	endpoint := c.BaseURL.JoinPath("zones")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result []Zone

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.SetBasicAuth(c.tokenID, c.tokenSecret)

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
