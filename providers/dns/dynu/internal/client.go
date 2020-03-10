package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-acme/lego/v3/log"
)

const defaultBaseURL = "https://api.dynu.com/v2"

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

func NewClient() *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    defaultBaseURL,
	}
}

// GetRecords Get DNS records based on a hostname and resource record type.
func (c Client) GetRecords(hostname string, recordType string) ([]DNSRecord, error) {
	endpoint, err := c.createEndpoint("dns", "record", hostname)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Set("recordType", recordType)
	endpoint.RawQuery = query.Encode()

	apiResp := RecordsResponse{}
	err = c.doRetry(http.MethodGet, endpoint.String(), nil, &apiResp)
	if err != nil {
		return nil, err
	}

	if apiResp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("API error: %w", apiResp.APIException)
	}

	return apiResp.DNSRecords, nil
}

// AddNewRecord Add a new DNS record for DNS service.
func (c Client) AddNewRecord(domainID int64, record DNSRecord) error {
	endpoint, err := c.createEndpoint("dns", strconv.FormatInt(domainID, 10), "record")
	if err != nil {
		return err
	}

	reqBody, err := json.Marshal(record)
	if err != nil {
		return err
	}

	apiResp := RecordResponse{}
	err = c.doRetry(http.MethodPost, endpoint.String(), reqBody, &apiResp)
	if err != nil {
		return err
	}

	if apiResp.StatusCode/100 != 2 {
		return fmt.Errorf("API error: %w", apiResp.APIException)
	}

	return nil
}

// DeleteRecord Remove a DNS record from DNS service.
func (c Client) DeleteRecord(domainID int64, recordID int64) error {
	endpoint, err := c.createEndpoint("dns", strconv.FormatInt(domainID, 10), "record", strconv.FormatInt(recordID, 10))
	if err != nil {
		return err
	}

	apiResp := APIException{}
	err = c.doRetry(http.MethodDelete, endpoint.String(), nil, &apiResp)
	if err != nil {
		return err
	}

	if apiResp.StatusCode/100 != 2 {
		return fmt.Errorf("API error: %w", apiResp)
	}

	return nil
}

// GetRootDomain Get the root domain name based on a hostname.
func (c Client) GetRootDomain(hostname string) (*DNSHostname, error) {
	endpoint, err := c.createEndpoint("dns", "getroot", hostname)
	if err != nil {
		return nil, err
	}

	apiResp := DNSHostname{}
	err = c.doRetry(http.MethodGet, endpoint.String(), nil, &apiResp)
	if err != nil {
		return nil, err
	}

	if apiResp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("API error: %w", apiResp.APIException)
	}

	return &apiResp, nil
}

// doRetry the API is really unstable so we need to retry on EOF.
func (c Client) doRetry(method, url string, body []byte, data interface{}) error {
	var resp *http.Response

	ctx, cancel := context.WithCancel(context.Background())

	operation := func() error {
		var reqBody io.Reader
		if len(body) > 0 {
			reqBody = bytes.NewReader(body)
		}

		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return err
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		resp, err = c.HTTPClient.Do(req)
		if errors.Is(err, io.EOF) {
			return err
		}

		if err != nil {
			cancel()
			return fmt.Errorf("client error: %w", err)
		}

		return nil
	}

	notify := func(err error, duration time.Duration) {
		log.Printf("client retries because of %v", err)
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 1 * time.Second

	err := backoff.RetryNotify(operation, backoff.WithContext(bo, ctx), notify)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return json.Unmarshal(all, data)
}

func (c Client) createEndpoint(fragments ...string) (*url.URL, error) {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	return baseURL.Parse(path.Join(baseURL.Path, path.Join(fragments...)))
}
