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
	"github.com/go-acme/lego/v5/internal/useragent"
)

const defaultBaseURL = "https://api.shellrent.com/api/v3"

// Client the Shellrent v3 API client.
type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(hc *http.Client) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{
		BaseURL:    baseURL,
		HTTPClient: hc,
	}
}

// ListPurchases lists all the purchases matching the search term.
// https://guide.shellrent.com/api/reference/list-purchases
func (c *Client) ListPurchases(ctx context.Context, search string) ([]Purchase, error) {
	var purchases []Purchase

	for page := 1; ; page++ {
		endpoint := c.BaseURL.JoinPath("purchases")

		query := endpoint.Query()
		query.Set("search", search)
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", "50")
		endpoint.RawQuery = query.Encode()

		req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}

		var result ListResponse[Purchase]

		err = c.do(req, &result)
		if err != nil {
			return nil, err
		}

		purchases = append(purchases, result.Data...)

		if page >= result.Meta.Pages {
			return purchases, nil
		}
	}
}

// GetDomain gets a domain.
// https://guide.shellrent.com/api/reference/get-domain
func (c *Client) GetDomain(ctx context.Context, domainID int) (*Domain, error) {
	endpoint := c.BaseURL.JoinPath("domains", strconv.Itoa(domainID))

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result APIResponse[*Domain]

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// CreateRecord creates a DNS record.
// https://guide.shellrent.com/api/reference/create-domain-dns-record
func (c *Client) CreateRecord(ctx context.Context, domainID int, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", strconv.Itoa(domainID), "dns", "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	var result APIResponse[*Record]

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// DeleteRecord deletes a DNS record.
// https://guide.shellrent.com/api/reference/delete-domain-dns-record
func (c *Client) DeleteRecord(ctx context.Context, domainID int, recordID string) error {
	endpoint := c.BaseURL.JoinPath("domains", strconv.Itoa(domainID), "dns", "records", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

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
