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
	querystring "github.com/google/go-querystring/query"
)

const AuthenticationHeader = "X-API-Key"

// Client the Poweradmin API client.
type Client struct {
	apiKey string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(baseURL, apiKey string) (*Client, error) {
	if baseURL == "" {
		return nil, errors.New("missing base URL")
	}

	if apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	bu, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    bu,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateRecord creates a new record in a zone.
// https://github.com/poweradmin/poweradmin/blob/c1d0a6f6c144f6b555766e6780c7dce40d072dc7/lib/Application/Controller/Api/V2/ZonesRecordsController.php#L365
func (c *Client) CreateRecord(ctx context.Context, zoneID int, record Record) (*Record, error) {
	endpoint := c.baseURL.JoinPath("api", "v2", "zones", strconv.Itoa(zoneID), "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse[RecordResponse])

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Data.Record, nil
}

// DeleteRecord deletes a record from a zone.
// Note: the endpoint returns a 204 (No content) but with a body, this is wrong.
// https://github.com/poweradmin/poweradmin/blob/c1d0a6f6c144f6b555766e6780c7dce40d072dc7/lib/Application/Controller/Api/V2/ZonesRecordsController.php#L876
func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordID int) error {
	endpoint := c.baseURL.JoinPath("api", "v2", "zones", strconv.Itoa(zoneID), "records", strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// GetZones list all zones accessible to the authenticated user.
// https://github.com/poweradmin/poweradmin/blob/c1d0a6f6c144f6b555766e6780c7dce40d072dc7/lib/Application/Controller/Api/V2/ZonesController.php#L118
func (c *Client) GetZones(ctx context.Context, pager *Pager) ([]Zone, *Pagination, error) {
	endpoint := c.baseURL.JoinPath("api", "v2", "zones")

	if pager != nil {
		values, err := querystring.Values(pager)
		if err != nil {
			return nil, nil, err
		}

		endpoint.RawQuery = values.Encode()
	}

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	result := new(APIResponse[ZonesResponse])

	err = c.do(req, result)
	if err != nil {
		return nil, nil, err
	}

	return result.Data.Zones, result.Pagination, nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set(AuthenticationHeader, c.apiKey)

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

	return errAPI
}
