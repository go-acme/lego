package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	querystring "github.com/google/go-querystring/query"
)

type Client struct {
	baseURL    *url.URL
	HTTPClient *http.Client

	username string
	password string
}

func NewClient(hostname, username, password string) *Client {
	baseURL, _ := url.Parse(fmt.Sprintf("https://%s/rest/", hostname))

	return &Client{
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    baseURL,
		username:   username,
		password:   password,
	}
}

func (c Client) ListRecords(ctx context.Context) ([]ResourceRecord, error) {
	endpoint := c.baseURL.JoinPath("dns_rr_list")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result []ResourceRecord

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c Client) GetRecord(ctx context.Context, id string) (*ResourceRecord, error) {
	endpoint := c.baseURL.JoinPath("dns_rr_info")

	query := endpoint.Query()
	query.Set("rr_id", id)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result []ResourceRecord

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return &result[0], nil
}

func (c Client) AddRecord(ctx context.Context, record ResourceRecord) (*BaseOutput, error) {
	endpoint := c.baseURL.JoinPath("dns_rr_add")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	var result []BaseOutput

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return &result[0], nil
}

func (c Client) DeleteRecord(ctx context.Context, params DeleteInputParameters) (*BaseOutput, error) {
	endpoint := c.baseURL.JoinPath("dns_rr_delete")

	// (rr_id || (rr_name && (dns_id || dns_name || hostaddr)))

	v, err := querystring.Values(params)
	if err != nil {
		return nil, fmt.Errorf("query parameters: %w", err)
	}
	endpoint.RawQuery = v.Encode()

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result []BaseOutput

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return &result[0], nil
}

func (c Client) do(req *http.Request, result any) error {
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("cache-control", "no-cache")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	switch req.Method {
	case http.MethodPost:
		if resp.StatusCode != http.StatusCreated {
			return parseError(req, resp)
		}
	default:
		if resp.StatusCode == http.StatusNoContent {
			return nil
		}

		if resp.StatusCode != http.StatusOK {
			return parseError(req, resp)
		}
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

	var response APIError
	err := json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code %d] %w", resp.StatusCode, response)
}
