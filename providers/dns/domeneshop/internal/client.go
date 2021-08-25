package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL string = "https://api.domeneshop.no/v0"

// Client implements a very simple wrapper around the Domeneshop API.
// For now it will only deal with adding and removing TXT records, as required by ACME providers.
// https://api.domeneshop.no/docs/
type Client struct {
	HTTPClient *http.Client
	baseURL    string
	apiToken   string
	apiSecret  string
}

// NewClient returns an instance of the Domeneshop API wrapper.
func NewClient(apiToken, apiSecret string) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    defaultBaseURL,
		apiToken:   apiToken,
		apiSecret:  apiSecret,
	}
}

// GetDomainByName fetches the domain list and returns the Domain object for the matching domain.
// https://api.domeneshop.no/docs/#operation/getDomains
func (c *Client) GetDomainByName(domain string) (*Domain, error) {
	var domains []Domain

	err := c.doRequest(http.MethodGet, "domains", nil, &domains)
	if err != nil {
		return nil, err
	}

	for _, d := range domains {
		if !d.Services.DNS {
			// Domains without DNS service cannot have DNS record added.
			continue
		}

		if d.Name == domain {
			return &d, nil
		}
	}

	return nil, fmt.Errorf("failed to find matching domain name: %s", domain)
}

// CreateTXTRecord creates a TXT record with the provided host (subdomain) and data.
// https://api.domeneshop.no/docs/#tag/dns/paths/~1domains~1{domainId}~1dns/post
func (c *Client) CreateTXTRecord(domain *Domain, host string, data string) error {
	jsonRecord, err := json.Marshal(DNSRecord{
		Data: data,
		Host: host,
		TTL:  300,
		Type: "TXT",
	})
	if err != nil {
		return err
	}

	return c.doRequest(http.MethodPost, fmt.Sprintf("domains/%d/dns", domain.ID), jsonRecord, nil)
}

// DeleteTXTRecord deletes the DNS record matching the provided host and data.
// https://api.domeneshop.no/docs/#tag/dns/paths/~1domains~1{domainId}~1dns~1{recordId}/delete
func (c *Client) DeleteTXTRecord(domain *Domain, host string, data string) error {
	record, err := c.getDNSRecordByHostData(*domain, host, data)
	if err != nil {
		return err
	}

	return c.doRequest(http.MethodDelete, fmt.Sprintf("domains/%d/dns/%d", domain.ID, record.ID), nil, nil)
}

// getDNSRecordByHostData finds the first matching DNS record with the provided host and data.
// https://api.domeneshop.no/docs/#operation/getDnsRecords
func (c *Client) getDNSRecordByHostData(domain Domain, host string, data string) (*DNSRecord, error) {
	var records []DNSRecord

	err := c.doRequest(http.MethodGet, fmt.Sprintf("domains/%d/dns", domain.ID), nil, &records)
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		if r.Host == host && r.Data == data {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("failed to find record with host %s for domain %s", host, domain.Name)
}

// doRequest makes a request against the API with an optional body,
// and makes sure that the required Authorization header is set using `setBasicAuth`.
func (c *Client) doRequest(method string, endpoint string, reqBody []byte, v interface{}) error {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", c.baseURL, endpoint), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.apiToken, c.apiSecret)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("API returned %s: %s", resp.Status, respBody)
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(&v)
	}

	return nil
}
