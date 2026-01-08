package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://api.selectel.ru/domains/v1"

const tokenHeader = "X-Token"

// Client represents the DNS client.
type Client struct {
	token string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient returns a client instance.
func NewClient(token string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		token:      token,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetDomainByName gets Domain object by its name. If `domainName` level > 2 and there is
// no such domain on the account - it'll recursively search for the first
// which is exists in Selectel Domain API.
func (c *Client) GetDomainByName(ctx context.Context, domainName string) (*Domain, error) {
	req, err := newJSONRequest(ctx, http.MethodGet, c.BaseURL.JoinPath(domainName), nil)
	if err != nil {
		return nil, err
	}

	domain := &Domain{}

	statusCode, err := c.do(req, domain)
	if err != nil {
		if statusCode == http.StatusNotFound && strings.Count(domainName, ".") > 1 {
			// Look up for the next subdomain
			_, after, _ := strings.Cut(domainName, ".")
			return c.GetDomainByName(ctx, after)
		}

		return nil, err
	}

	return domain, nil
}

// AddRecord adds Record for given domain.
func (c *Client) AddRecord(ctx context.Context, domainID int, body Record) (*Record, error) {
	req, err := newJSONRequest(ctx, http.MethodPost, c.BaseURL.JoinPath(strconv.Itoa(domainID), "records", "/"), body)
	if err != nil {
		return nil, err
	}

	record := &Record{}

	_, err = c.do(req, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// ListRecords returns list records for specific domain.
func (c *Client) ListRecords(ctx context.Context, domainID int) ([]Record, error) {
	req, err := newJSONRequest(ctx, http.MethodGet, c.BaseURL.JoinPath(strconv.Itoa(domainID), "records", "/"), nil)
	if err != nil {
		return nil, err
	}

	var records []Record

	_, err = c.do(req, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// DeleteRecord deletes specific record.
func (c *Client) DeleteRecord(ctx context.Context, domainID, recordID int) error {
	endpoint := c.BaseURL.JoinPath(strconv.Itoa(domainID), "records", strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)

	return err
}

func (c *Client) do(req *http.Request, result any) (int, error) {
	req.Header.Set(tokenHeader, c.token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return resp.StatusCode, parseError(req, resp)
	}

	if result == nil {
		return resp.StatusCode, nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return resp.StatusCode, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return resp.StatusCode, nil
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

	errAPI := &APIError{}

	err := json.Unmarshal(raw, errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("request failed with status code %d: %w", resp.StatusCode, errAPI)
}
