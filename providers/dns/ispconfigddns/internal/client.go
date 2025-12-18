package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

const (
	addAction    = "add"
	deleteAction = "delete"
)

type Client struct {
	token string

	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(serverURL, token string) (*Client, error) {
	_, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("server URL: %w", err)
	}

	return &Client{
		BaseURL:    serverURL,
		token:      token,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddTXTRecord(ctx context.Context, zone, fqdn, content string) error {
	return c.updateRecord(ctx, UpdateRecord{Action: addAction, Zone: zone, Type: "TXT", Record: fqdn, Data: content})
}

func (c *Client) DeleteTXTRecord(ctx context.Context, zone, fqdn, recordContent string) error {
	return c.updateRecord(ctx, UpdateRecord{Action: deleteAction, Zone: zone, Type: "TXT", Record: fqdn, Data: recordContent})
}

func (c *Client) updateRecord(ctx context.Context, action UpdateRecord) error {
	req, err := c.newRequest(ctx, action)
	if err != nil {
		return err
	}

	return c.do(req)
}

func (c *Client) do(req *http.Request) error {
	useragent.SetHeader(req.Header)

	req.SetBasicAuth("anonymous", c.token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return nil
}

func (c *Client) newRequest(ctx context.Context, action UpdateRecord) (*http.Request, error) {
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	endpoint = endpoint.JoinPath("ddns", "update.php")

	values, err := querystring.Values(action)
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = values.Encode()

	method := http.MethodPost
	if action.Action == deleteAction {
		method = http.MethodDelete
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}
