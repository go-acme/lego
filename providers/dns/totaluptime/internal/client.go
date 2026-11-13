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
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://api.totaluptime.com"

// Client the Total Uptime API client.
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

// AddTXTRecord adds a record to an existing domain.
// https://totaluptime.com/api/v2/dns/add-record-existing-domain/
// https://api.totaluptime.com/api-docs/index.html#!/Cloud32DNS32Calls/AddRecordType
func (c *Client) AddTXTRecord(ctx context.Context, domainID string, record TXTRecord) (*APIResponse, error) {
	endpoint := c.BaseURL.JoinPath("CloudDNS", "Domain", domainID, "TXTRecord")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteTXTRecord deletes an existing resource record.
// https://totaluptime.com/api/v2/dns/delete-existing-resource-record/
// https://api.totaluptime.com/api-docs/index.html#!/Cloud32DNS32Calls/DeleteRecord
func (c *Client) DeleteTXTRecord(ctx context.Context, domainID, recordID string) (*APIResponse, error) {
	endpoint := c.BaseURL.JoinPath("CloudDNS", "Domain", domainID, "TXTRecord", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetDomains returns domains within the account, including all sub-companies.
// https://totaluptime.com/api/v2/dns/retrieve-all-domains/
// https://api.totaluptime.com/api-docs/index.html#!/Cloud32DNS32Calls/GetAllDomain
func (c *Client) GetDomains(ctx context.Context, options Options) ([]Domain, error) {
	endpoint := c.BaseURL.JoinPath("CloudDNS", "Domain", "ALL")

	values, err := querystring.Values(options)
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = values.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result DomainsResponse

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Rows, nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.SetBasicAuth(c.username, c.password)

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
