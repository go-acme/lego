package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
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

// GetTxtRecord gets a TXT record.
func (c *Client) GetTxtRecord(domain, name, value string) (*DNSRecord, error) {
	records, err := c.getRecords(domain, name)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if equalsTXTRecord(record, name, value) {
			return &record, nil
		}
	}

	return nil, fmt.Errorf("could not find record: Domain: %s; Record: %s", domain, name)
}

// https://www.arvancloud.com/docs/api/cdn/4.0#operation/dns_records.list
func (c *Client) getRecords(domain, search string) ([]DNSRecord, error) {
	endpoint, err := c.createEndpoint("cdn", "4.0", "domains", domain, "dns-records")
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	if search != "" {
		query := endpoint.Query()
		query.Set("search", strings.ReplaceAll(search, "_", ""))
		endpoint.RawQuery = query.Encode()
	}

	resp, err := c.do(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not get records %s: Domain: %s; Status: %s; Body: %s",
			search, domain, resp.Status, string(body))
	}

	response := &apiResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	var records []DNSRecord
	err = json.Unmarshal(response.Data, &records)
	if err != nil {
		return nil, fmt.Errorf("failed to decode records: %w", err)
	}

	return records, nil
}

// CreateRecord creates a DNS record.
// https://www.arvancloud.com/docs/api/cdn/4.0#operation/dns_records.create
func (c *Client) CreateRecord(domain string, record DNSRecord) (*DNSRecord, error) {
	reqBody, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	endpoint, err := c.createEndpoint("cdn", "4.0", "domains", domain, "dns-records")
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	resp, err := c.do(http.MethodPost, endpoint.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("could not create record %s; Domain: %s; Status: %s; Body: %s", string(reqBody), domain, resp.Status, string(body))
	}

	response := &apiResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	var newRecord DNSRecord
	err = json.Unmarshal(response.Data, &newRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to decode record: %w", err)
	}

	return &newRecord, nil
}

// DeleteRecord deletes a DNS record.
// https://www.arvancloud.com/docs/api/cdn/4.0#operation/dns_records.remove
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("could not delete record %s; Domain: %s; Status: %s; Body: %s", id, domain, resp.Status, string(body))
	}

	return nil
}

func (c *Client) do(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
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

func equalsTXTRecord(record DNSRecord, name, value string) bool {
	if record.Type != "txt" {
		return false
	}

	if record.Name != name {
		return false
	}

	data, ok := record.Value.(map[string]interface{})
	if !ok {
		return false
	}

	return data["text"] == value
}
