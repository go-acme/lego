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
const defaultBaseURL = "https://napi.arvancloud.com/cdn/4.0/"

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
	records, err := c.getRecords(domain)
	if err != nil {
		return nil, err
	}

	for _, record := range records.Records {
		if record.Type == "txt" && record.Name == name && equalValue(record.Value, value) {
			return &record, nil
		}
	}

	return nil, fmt.Errorf("could not find record: domain name: %s; Record: %s", domain, name)
}

// https://napi.arvancloud.com/cdn/4.0/domains/{domain}/dns-records
func (c *Client) getRecords(domain string) (*DNSRecords, error) {
	endpoint, err := c.createEndpoint("domains", domain, "dns-records")
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	resp, err := c.do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("could not get records: domain Name: %s; Status: %s; Body: %s",
			domain, resp.Status, string(bodyBytes))
	}

	records := &DNSRecords{}
	err = json.NewDecoder(resp.Body).Decode(records)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return records, nil
}

// CreateRecord creates a DNS record.
// https://napi.arvancloud.com/cdn/4.0/domains/{domain}/dns-records
func (c *Client) CreateRecord(domain string, record DNSRecord) error {
	body, err := json.Marshal(record)
	if err != nil {
		return err
	}

	endpoint, err := c.createEndpoint("domains", domain, "dns-records")
	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	resp, err := c.do(http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("could not create record %s; Status: %s; Body: %s", string(body), resp.Status, string(bodyBytes))
	}

	return nil
}

// DeleteRecord deletes a DNS record.
// https://napi.arvancloud.com/cdn/4.0/domains/{domain}/dns-records/{id}
func (c *Client) DeleteRecord(domain string, recordID string) error {
	endpoint, err := c.createEndpoint("domains", domain, "dns-records", recordID)
	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	resp, err := c.do(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	fmt.Println(resp)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not delete record: %s; Status: %s", resp.Status, recordID)
	}

	return nil
}

func (c *Client) do(method string, endpoint fmt.Stringer, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if method != http.MethodDelete {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set(authHeader, c.apiKey)

	return c.HTTPClient.Do(req)
}

func (c *Client) createEndpoint(parts ...string) (*url.URL, error) {
	//baseURL, err := url.Parse("https://napi.arvancloud.com/cdn/4.0/")
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

func equalValue(in interface{}, value string) bool {
	v, ok := in.(map[string]interface{})
	if !ok {
		return false
	}
	for key, val := range v {
		if key == "text" && val == value {
			return true
		}
	}
	return false
}
