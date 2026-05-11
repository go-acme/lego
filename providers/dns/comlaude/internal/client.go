package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://api.comlaude.com"

// Client the Com Laude API client.
type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient() (*Client, error) {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) GetDomains(ctx context.Context, groupID, name string, pager *Pager) (*DomainsResponse, error) {
	endpoint := c.BaseURL.JoinPath("groups", groupID, "domains")

	values, err := querystring.Values(pager)
	if err != nil {
		return nil, err
	}

	values.Set("filter[name]", name)

	endpoint.RawQuery = values.Encode()

	req, err := newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(DomainsResponse)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) CreateRecord(ctx context.Context, groupID, zoneID string, record RecordRequest) (string, error) {
	endpoint := c.BaseURL.JoinPath("groups", groupID, "zones", zoneID, "records")

	req, err := newRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return "", err
	}

	result := new(RecordResponse)

	err = c.do(req, result)
	if err != nil {
		return "", err
	}

	return result.Data.ID, nil
}

func (c *Client) DeleteRecord(ctx context.Context, groupID, zoneID, recordID string) error {
	endpoint := c.BaseURL.JoinPath("groups", groupID, "zones", zoneID, "records", recordID)

	req, err := newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

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

func newRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	values := url.Values{}

	if payload != nil {
		var err error

		values, err = querystring.Values(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	tok := getToken(ctx)
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
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
