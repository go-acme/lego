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
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

const defaultBaseURL = "https://usersapiv2.epik.com/v2"

// Client the Epik API client.
type Client struct {
	signature string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(signature string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		signature:  signature,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetDNSRecords gets DNS records for a domain.
// https://docs.userapi.epik.com/v2/#/DNS%20Host%20Records/getDnsRecord
func (c *Client) GetDNSRecords(ctx context.Context, domain string) ([]Record, error) {
	endpoint := c.createEndpoint(domain, url.Values{})

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var data GetDNSRecordResponse

	err = c.do(req, &data)
	if err != nil {
		return nil, err
	}

	return data.Data.Records, nil
}

// CreateHostRecord creates a record for a domain.
// https://docs.userapi.epik.com/v2/#/DNS%20Host%20Records/createHostRecord
func (c *Client) CreateHostRecord(ctx context.Context, domain string, record RecordRequest) (*Data, error) {
	endpoint := c.createEndpoint(domain, url.Values{})

	payload := CreateHostRecords{Payload: record}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}

	var data Data

	err = c.do(req, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// RemoveHostRecord removes a record for a domain.
// https://docs.userapi.epik.com/v2/#/DNS%20Host%20Records/removeHostRecord
func (c *Client) RemoveHostRecord(ctx context.Context, domain, recordID string) (*Data, error) {
	params := url.Values{}
	params.Set("ID", recordID)

	endpoint := c.createEndpoint(domain, params)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var data Data

	err = c.do(req, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
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

func (c *Client) createEndpoint(domain string, params url.Values) *url.URL {
	endpoint := c.baseURL.JoinPath("domains", domain, "records")

	params.Set("SIGNATURE", c.signature)
	endpoint.RawQuery = params.Encode()

	return endpoint
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

	var apiErr APIError

	err := json.Unmarshal(raw, &apiErr)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return &apiErr
}
