package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://support.euserv.com"

// Client the EUserv API client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient() *Client {
	return &Client{
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetRecords gets a data list of DNS record data for all domains in an order.
// https://support.euserv.com/api-doc/#api-Domain-kc2_domain_dns_get_records
func (c *Client) GetRecords(ctx context.Context, request GetRecordsRequest) ([]Domain, error) {
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Set("subaction", "kc2_domain_dns_get_records")
	endpoint.RawQuery = query.Encode()

	req, err := newHTTPRequest(ctx, endpoint, request)
	if err != nil {
		return nil, err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return nil, err
	}

	result, err := extractResponse[Records](response)
	if err != nil {
		return nil, err
	}

	return result.Domains, nil
}

// RemoveRecord removes a DNS record.
// https://support.euserv.com/api-doc/#api-Domain-kc2_domain_dns_remove
func (c *Client) RemoveRecord(ctx context.Context, recordID string) error {
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	query := endpoint.Query()
	query.Set("subaction", "kc2_domain_dns_remove")
	query.Set("dns_record_id", recordID)
	endpoint.RawQuery = query.Encode()

	req, err := newHTTPRequest(ctx, endpoint, nil)
	if err != nil {
		return err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return err
	}

	_, err = extractResponse[Session](response)
	if err != nil {
		return err
	}

	return nil
}

// SetRecord create or updates a DNS record.
// https://support.euserv.com/api-doc/#api-Domain-kc2_domain_dns_set
func (c *Client) SetRecord(ctx context.Context, request SetRecordRequest) error {
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	query := endpoint.Query()
	query.Set("subaction", "kc2_domain_dns_set")
	endpoint.RawQuery = query.Encode()

	req, err := newHTTPRequest(ctx, endpoint, request)
	if err != nil {
		return err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return err
	}

	_, err = extractResponse[Session](response)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
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

func newHTTPRequest(ctx context.Context, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	query := endpoint.Query()
	query.Set("method", "json")
	query.Set("lang_id", "2")

	sessionID := getSessionID(ctx)
	if sessionID != "" {
		query.Set("sess_id", sessionID)
	}

	if payload != nil {
		values, err := querystring.Values(payload)
		if err != nil {
			return nil, err
		}

		maps.Copy(query, values)
	}

	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func extractResponse[T any](response APIResponse) (T, error) {
	if response.Code != "100" {
		var zero T

		return zero, &APIError{APIResponse: response}
	}

	var result T

	err := json.Unmarshal(response.Result, &result)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("unable to unmarshal response: %s, %w", string(response.Result), err)
	}

	return result, nil
}
