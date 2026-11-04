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
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
)

const defaultBaseURL = "https://www.lima-city.de/usercp"

type Client struct {
	apiKey     string
	baseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(apiKey string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetDomains retrieves a list of all domains owned by the authenticated user.
// https://www.lima-city.de/docs#tag/domains/operation/listDomains
func (c *Client) GetDomains(ctx context.Context) ([]Domain, error) {
	endpoint := c.baseURL.JoinPath("domains.json")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var results DomainsResponse

	err = c.do(req, &results)
	if err != nil {
		return nil, err
	}

	return results.Data, nil
}

// GetRecords retrieves all DNS records configured for a domain.
// https://www.lima-city.de/docs#tag/dns/operation/listDnsRecords
func (c *Client) GetRecords(ctx context.Context, domainID int) ([]Record, error) {
	endpoint := c.baseURL.JoinPath("domains", strconv.Itoa(domainID), "records.json")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var results RecordsResponse

	err = c.do(req, &results)
	if err != nil {
		return nil, err
	}

	return results.Data, nil
}

// AddRecord creates a new DNS resource record for the domain.
// https://www.lima-city.de/docs#tag/dns/operation/createDnsRecord
func (c *Client) AddRecord(ctx context.Context, domainID int, record Record) error {
	endpoint := c.baseURL.JoinPath("domains", strconv.Itoa(domainID), "records.json")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, NameserverRecordPayload{Data: record})
	if err != nil {
		return err
	}

	var results APIResponse

	return c.do(req, &results)
}

// UpdateRecord updates an existing DNS resource record.
// https://www.lima-city.de/docs#tag/dns/operation/updateDnsRecord
func (c *Client) UpdateRecord(ctx context.Context, domainID, recordID int, record Record) error {
	endpoint := c.baseURL.JoinPath("domains", strconv.Itoa(domainID), "records", strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, NameserverRecordPayload{Data: record})
	if err != nil {
		return err
	}

	var results APIResponse

	return c.do(req, &results)
}

// DeleteRecord deletes an existing DNS resource record.
// https://www.lima-city.de/docs#tag/dns/operation/deleteDnsRecord
func (c *Client) DeleteRecord(ctx context.Context, domainID, recordID int) error {
	endpoint := c.baseURL.JoinPath("domains", strconv.Itoa(domainID), "records", strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	var results APIResponse

	return c.do(req, &results)
}

func (c *Client) do(req *http.Request, result any) error {
	req.SetBasicAuth("api", c.apiKey)

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

	errAPI := new(APIResponse)

	err := json.Unmarshal(raw, errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code: %d] %w", resp.StatusCode, errAPI)
}
