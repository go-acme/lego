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
)

const defaultBaseURL = "https://www.webnames.ca/_/APICore"

// Client the webnames.ca API client.
type Client struct {
	user string
	key  string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(user, key string) (*Client, error) {
	if user == "" || key == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		user:       user,
		key:        key,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddTXTRecord(ctx context.Context, domainName, hostName, value string) ([]DNSRecordSet, error) {
	endpoint := c.BaseURL.JoinPath("domains", domainName, "add-txt-record")

	query := endpoint.Query()
	query.Set("hostName", hostName)
	query.Set("txt", value)

	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result APIResponse[*DNSInfo]

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Result.DNSRecordSets, nil
}

func (c *Client) DeleteTXTRecord(ctx context.Context, domainName, hostName, value string) ([]DNSRecordSet, error) {
	endpoint := c.BaseURL.JoinPath("domains", domainName, "delete-txt-record")

	query := endpoint.Query()
	query.Set("hostName", hostName)
	query.Set("txt", value)

	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result APIResponse[*DNSInfo]

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Result.DNSRecordSets, nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set("API-User", c.user)
	req.Header.Set("API-Key", c.key)

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
