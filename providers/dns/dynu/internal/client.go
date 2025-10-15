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
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/wait"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://api.dynu.com/v2"

type Client struct {
	baseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient() *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    baseURL,
	}
}

// GetRecords Get DNS records based on a hostname and resource record type.
func (c *Client) GetRecords(ctx context.Context, hostname, recordType string) ([]DNSRecord, error) {
	endpoint := c.baseURL.JoinPath("dns", "record", hostname)

	query := endpoint.Query()
	query.Set("recordType", recordType)
	endpoint.RawQuery = query.Encode()

	apiResp := RecordsResponse{}
	err := c.doRetry(ctx, http.MethodGet, endpoint.String(), nil, &apiResp)
	if err != nil {
		return nil, err
	}

	if apiResp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("API error: %w", apiResp.APIException)
	}

	return apiResp.DNSRecords, nil
}

// AddNewRecord Add a new DNS record for DNS service.
func (c *Client) AddNewRecord(ctx context.Context, domainID int64, record DNSRecord) error {
	endpoint := c.baseURL.JoinPath("dns", strconv.FormatInt(domainID, 10), "record")

	reqBody, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to create request JSON body: %w", err)
	}

	apiResp := RecordResponse{}
	err = c.doRetry(ctx, http.MethodPost, endpoint.String(), reqBody, &apiResp)
	if err != nil {
		return err
	}

	if apiResp.StatusCode/100 != 2 {
		return fmt.Errorf("API error: %w", apiResp.APIException)
	}

	return nil
}

// DeleteRecord Remove a DNS record from DNS service.
func (c *Client) DeleteRecord(ctx context.Context, domainID, recordID int64) error {
	endpoint := c.baseURL.JoinPath("dns", strconv.FormatInt(domainID, 10), "record", strconv.FormatInt(recordID, 10))

	apiResp := APIException{}
	err := c.doRetry(ctx, http.MethodDelete, endpoint.String(), nil, &apiResp)
	if err != nil {
		return err
	}

	if apiResp.StatusCode/100 != 2 {
		return fmt.Errorf("API error: %w", apiResp)
	}

	return nil
}

// GetRootDomain Get the root domain name based on a hostname.
func (c *Client) GetRootDomain(ctx context.Context, hostname string) (*DNSHostname, error) {
	endpoint := c.baseURL.JoinPath("dns", "getroot", hostname)

	apiResp := DNSHostname{}
	err := c.doRetry(ctx, http.MethodGet, endpoint.String(), nil, &apiResp)
	if err != nil {
		return nil, err
	}

	if apiResp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("API error: %w", apiResp.APIException)
	}

	return &apiResp, nil
}

// doRetry the API is really unstable, so we need to retry on EOF.
func (c *Client) doRetry(ctx context.Context, method, uri string, body []byte, result any) error {
	operation := func() error {
		return c.do(ctx, method, uri, body, result)
	}

	notify := func(err error, duration time.Duration) {
		log.Printf("client retries because of %v", err)
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 1 * time.Second

	return wait.Retry(ctx, operation, backoff.WithBackOff(bo), backoff.WithNotify(notify))
}

func (c *Client) do(ctx context.Context, method, uri string, body []byte, result any) error {
	var reqBody io.Reader
	if len(body) > 0 {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, reqBody)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if errors.Is(err, io.EOF) {
		return err
	}

	if err != nil {
		return backoff.Permanent(fmt.Errorf("client error: %w", errutils.NewHTTPDoError(req, err)))
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return backoff.Permanent(errutils.NewReadResponseError(req, resp.StatusCode, err))
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return backoff.Permanent(errutils.NewUnmarshalError(req, resp.StatusCode, raw, err))
	}

	return nil
}
