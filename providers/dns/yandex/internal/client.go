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
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://pddimp.yandex.ru/api2/admin/dns"

const successCode = "ok"

const pddTokenHeader = "PddToken"

type Client struct {
	pddToken string

	baseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(pddToken string) (*Client, error) {
	if pddToken == "" {
		return nil, errors.New("PDD token is required")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		pddToken:   pddToken,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddRecord(ctx context.Context, payload Record) (*Record, error) {
	endpoint := c.baseURL.JoinPath("add")

	req, err := newRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}

	r := AddResponse{}

	err = c.do(req, &r)
	if err != nil {
		return nil, err
	}

	return r.Record, nil
}

func (c *Client) RemoveRecord(ctx context.Context, payload Record) (int, error) {
	endpoint := c.baseURL.JoinPath("del")

	req, err := newRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return 0, err
	}

	r := RemoveResponse{}

	err = c.do(req, &r)
	if err != nil {
		return 0, err
	}

	return r.RecordID, nil
}

func (c *Client) GetRecords(ctx context.Context, domain string) ([]Record, error) {
	endpoint := c.baseURL.JoinPath("list")

	payload := struct {
		Domain string `url:"domain"`
	}{Domain: domain}

	req, err := newRequest(ctx, http.MethodGet, endpoint, payload)
	if err != nil {
		return nil, err
	}

	r := ListResponse{}

	err = c.do(req, &r)
	if err != nil {
		return nil, err
	}

	return r.Records, nil
}

func (c *Client) do(req *http.Request, result Response) error {
	req.Header.Set(pddTokenHeader, c.pddToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if result.GetSuccess() != successCode {
		return fmt.Errorf("error during operation: %s %s", result.GetSuccess(), result.GetError())
	}

	return nil
}

func newRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		switch method {
		case http.MethodPost:
			values, err := querystring.Values(payload)
			if err != nil {
				return nil, err
			}

			buf.WriteString(values.Encode())

		case http.MethodGet:
			values, err := querystring.Values(payload)
			if err != nil {
				return nil, err
			}

			endpoint.RawQuery = values.Encode()
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return req, nil
}
