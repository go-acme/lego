package internal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	querystring "github.com/google/go-querystring/query"
)

const statusSuccess = "success"

const defaultBaseURL = "https://my.axelname.ru/rest/"

// Client the Axelname API client.
type Client struct {
	nickname string
	token    string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(nickname, token string) (*Client, error) {
	if token == "" || nickname == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		nickname:   nickname,
		token:      token,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) ListRecords(ctx context.Context, domain string) ([]Record, error) {
	endpoint := c.baseURL.JoinPath("dns_list")

	query := endpoint.Query()
	query.Set("domain", domain)

	endpoint.RawQuery = query.Encode()

	req, err := c.newRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var results ListResponse

	err = c.do(req, &results)
	if err != nil {
		return nil, err
	}

	if results.Result != statusSuccess {
		return nil, &results.APIError
	}

	return results.List, nil
}

func (c *Client) DeleteRecord(ctx context.Context, domain string, record Record) error {
	endpoint := c.baseURL.JoinPath("dns_delete")

	values, err := querystring.Values(record)
	if err != nil {
		return err
	}

	values.Set("domain", domain)

	endpoint.RawQuery = values.Encode()

	req, err := c.newRequest(ctx, endpoint)
	if err != nil {
		return err
	}

	var results APIResponse

	err = c.do(req, &results)
	if err != nil {
		return err
	}

	if results.Result != statusSuccess {
		return &results.APIError
	}

	return nil
}

func (c *Client) AddRecord(ctx context.Context, domain string, record Record) error {
	endpoint := c.baseURL.JoinPath("dns_add")

	values, err := querystring.Values(record)
	if err != nil {
		return err
	}

	values.Set("domain", domain)

	endpoint.RawQuery = values.Encode()

	req, err := c.newRequest(ctx, endpoint)
	if err != nil {
		return err
	}

	var results APIResponse

	err = c.do(req, &results)
	if err != nil {
		return err
	}

	if results.Result != statusSuccess {
		return &results.APIError
	}

	return nil
}

func (c *Client) newRequest(ctx context.Context, endpoint *url.URL) (*http.Request, error) {
	query := endpoint.Query()
	query.Set("token", c.token)
	query.Set("nichdl", c.nickname)

	endpoint.RawQuery = query.Encode()

	return http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
}

func (c *Client) do(req *http.Request, result any) error {
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

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	var errAPI APIError

	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return &errAPI
}
