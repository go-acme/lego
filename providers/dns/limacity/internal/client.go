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

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
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

func (c *Client) AddRecord(ctx context.Context, domainID int, record Record) error {
	endpoint := c.baseURL.JoinPath("domains", strconv.Itoa(domainID), "records.json")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, NameserverRecordPayload{Data: record})
	if err != nil {
		return err
	}

	var results APIResponse
	err = c.do(req, &results)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateRecord(ctx context.Context, domainID, recordID int, record Record) error {
	endpoint := c.baseURL.JoinPath("domains", strconv.Itoa(domainID), "records", strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, NameserverRecordPayload{Data: record})
	if err != nil {
		return err
	}

	var results APIResponse
	err = c.do(req, &results)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteRecord(ctx context.Context, domainID, recordID int) error {
	// /domains/{domainId}/records/{recordId} DELETE
	endpoint := c.baseURL.JoinPath("domains", strconv.Itoa(domainID), "records", strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	var results APIResponse
	err = c.do(req, &results)
	if err != nil {
		return err
	}

	return nil
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

	var errAPI APIResponse
	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code: %d] %w", resp.StatusCode, &errAPI)
}
