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
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://api.openprovider.eu/v1beta/"

// Client the Openprovider API client.
type Client struct {
	username string
	password string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(username, password string) (*Client, error) {
	if username == "" || password == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		username:   username,
		password:   password,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// ListZones lists the zones.
// https://docs.openprovider.com/doc/all#operation/ListZones
func (c *Client) ListZones(ctx context.Context, zr *ZonesRequest) ([]Zone, error) {
	endpoint := c.BaseURL.JoinPath("dns", "zones")

	values, err := querystring.Values(zr)
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = values.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	results := &APIResponse[Results[Zone]]{}

	err = c.do(req, results)
	if err != nil {
		return nil, err
	}

	return results.Data.Results, nil
}

// UpdateZone updates a zone.
// https://docs.openprovider.com/doc/all#operation/UpdateZone
func (c *Client) UpdateZone(ctx context.Context, domain string, action ZoneAction) error {
	endpoint := c.BaseURL.JoinPath("dns", "zones", domain)

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, action)
	if err != nil {
		return err
	}

	results := &APIResponse[ResponseSuccess]{}

	err = c.do(req, results)
	if err != nil {
		return err
	}

	if !results.Data.Success {
		return fmt.Errorf("update failure (code: %d): %s", results.Code, results.Desc)
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	at := getToken(req.Context())
	if at != "" {
		req.Header.Set(authorizationHeader, "Bearer "+at)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode > http.StatusBadRequest {
		return parseError(req, resp)
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

	return fmt.Errorf("[status code %d] %w", resp.StatusCode, &errAPI)
}
