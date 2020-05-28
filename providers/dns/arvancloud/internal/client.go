package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

// defaultBaseURL represents the API endpoint to call.
const defaultBaseURL = "https://napi.arvancloud.com"

const authHeader = "Authorization"

// Client the ArvanCloud client.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string

	apiKey string
}

// NewClient Creates a new ArvanCloud client.
func NewClient(apiKey string) *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    defaultBaseURL,
		apiKey:     apiKey,
	}
}

// TxtRecord gets a TXT record.
func (c *Client) TxtRecord(domain, name, value string) (*DNSRecord, error) {
	records, err := c.records(domain, value)
	if err != nil {
		return nil, err
	}

	if len(records.Records) == 1 {
		record := records.Records[0]
		return &record, nil
	}

	return nil, fmt.Errorf("could not find record: Domain: %s; Record: %s", domain, name)
}

func (c *Client) records(domain, search string) (*DNSRecords, error) {
	endpoint, err := c.createEndpoint("cdn", "4.0", "domains", domain, "dns-records")
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	query := endpoint.Query()
	query.Set("search", search)
	endpoint.RawQuery = query.Encode()

	resp, err := c.do(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("could not get records %s: Domain: %s; Status: %s; Body: %s",
			search, domain, resp.Status, string(bodyBytes))
	}

	records := &DNSRecords{}
	err = json.NewDecoder(resp.Body).Decode(records)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return records, nil
}

// CreateRecord creates a DNS record.
func (c *Client) CreateRecord(domain string, record DNSRecord) error {
	body, err := json.Marshal(record)
	if err != nil {
		return err
	}

	endpoint, err := c.createEndpoint("cdn", "4.0", "domains", domain, "dns-records")
	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	resp, err := c.do(http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("could not create record %s; Domain: %s; Status: %s; Body: %s", string(body), domain, resp.Status, string(bodyBytes))
	}

	return nil
}

// DeleteRecord deletes a DNS record.
func (c *Client) DeleteRecord(domain, id string) error {
	endpoint, err := c.createEndpoint("cdn", "4.0", "domains", domain, "dns-records", id)
	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	resp, err := c.do(http.MethodDelete, endpoint.String(), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not delete record %s; Domain: %s; Status: %s", id, domain, resp.Status)
	}

	return nil
}

func (c *Client) do(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(authHeader, c.apiKey)

	return c.HTTPClient.Do(req)
}

func (c *Client) createEndpoint(parts ...string) (*url.URL, error) {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	endpoint, err := baseURL.Parse(path.Join(parts...))
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}
