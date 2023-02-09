package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://usersapiv2.epik.com/v2"

type Client struct {
	HTTPClient *http.Client
	baseURL    *url.URL
	signature  string
}

func NewClient(signature string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    baseURL,
		signature:  signature,
	}
}

// GetDNSRecords gets DNS records for a domain.
// https://docs.userapi.epik.com/v2/#/DNS%20Host%20Records/getDnsRecord
func (c Client) GetDNSRecords(domain string) ([]Record, error) {
	resp, err := c.do(http.MethodGet, domain, url.Values{}, nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body (%d): %w", resp.StatusCode, err)
	}

	err = checkError(resp.StatusCode, all)
	if err != nil {
		return nil, err
	}

	var data GetDNSRecordResponse
	err = json.Unmarshal(all, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request body (%d): %s", resp.StatusCode, string(all))
	}

	return data.Data.Records, nil
}

// CreateHostRecord creates a record for a domain.
// https://docs.userapi.epik.com/v2/#/DNS%20Host%20Records/createHostRecord
func (c Client) CreateHostRecord(domain string, record RecordRequest) (*Data, error) {
	payload := CreateHostRecords{Payload: record}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(http.MethodPost, domain, url.Values{}, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body (%d): %w", resp.StatusCode, err)
	}

	err = checkError(resp.StatusCode, all)
	if err != nil {
		return nil, err
	}

	var data Data
	err = json.Unmarshal(all, &data)
	if err != nil {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, string(all))
	}

	return &data, nil
}

// RemoveHostRecord removes a record for a domain.
// https://docs.userapi.epik.com/v2/#/DNS%20Host%20Records/removeHostRecord
func (c Client) RemoveHostRecord(domain string, recordID string) (*Data, error) {
	params := url.Values{}
	params.Set("ID", recordID)

	resp, err := c.do(http.MethodDelete, domain, params, nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body (%d): %w", resp.StatusCode, err)
	}

	err = checkError(resp.StatusCode, all)
	if err != nil {
		return nil, err
	}

	var data Data
	err = json.Unmarshal(all, &data)
	if err != nil {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, string(all))
	}

	return &data, nil
}

func (c *Client) do(method, domain string, params url.Values, body io.Reader) (*http.Response, error) {
	endpoint := c.baseURL.JoinPath("domains", domain, "records")

	params.Set("SIGNATURE", c.signature)
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.HTTPClient.Do(req)
}

func checkError(statusCode int, all []byte) error {
	if statusCode == http.StatusOK {
		return nil
	}

	var apiErr APIError
	err := json.Unmarshal(all, &apiErr)
	if err != nil {
		return fmt.Errorf("%d: %s", statusCode, string(all))
	}

	return &apiErr
}
