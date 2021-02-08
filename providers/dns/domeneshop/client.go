package domeneshop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const apiURL string = "https://api.domeneshop.no/v0"

// Client implements a very simple wrapper around the Domeneshop API.
//
// For now it will only deal with adding and removing TXT records, as
// required by ACME providers.
//
// https://api.domeneshop.no/docs/
type Client struct {
	APIToken   string
	APISecret  string
	HTTPClient *http.Client
}

// Domain JSON data structure
type Domain struct {
	Name           string   `json:"domain"`
	ID             int      `json:"id"`
	ExpiryDate     string   `json:"expiry_date"`
	Nameservers    []string `json:"nameservers"`
	RegisteredDate string   `json:"registered_date"`
	Registrant     string   `json:"registrant"`
	Renew          bool     `json:"renew"`
	Services       struct {
		DNS       bool   `json:"dns"`
		Email     bool   `json:"email"`
		Registrar bool   `json:"registrar"`
		Webhotel  string `json:"webhotel"`
	} `json:"services"`
	Status string
}

// DNSRecord JSON data structure
type DNSRecord struct {
	Data string `json:"data"`
	Host string `json:"host"`
	ID   int    `json:"id"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
}

// NewClient returns an instance of the Domeneshop API wrapper
func NewClient(apiToken, apiSecret string, httpClient *http.Client) *Client {
	client := Client{
		APIToken:   apiToken,
		APISecret:  apiSecret,
		HTTPClient: httpClient,
	}

	return &client
}

// Request makes a request against the API with an optional body, and makes sure
// that the required Authorization header is set using `setBasicAuth`
func (c *Client) Request(method string, endpoint string, reqBody []byte, v interface{}) error {

	var buf = bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", apiURL, endpoint), buf)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.APIToken, c.APISecret)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode > 399 {
		return fmt.Errorf("API returned %s: %s", resp.Status, respBody)
	}

	if v != nil {
		return json.Unmarshal(respBody, &v)
	}
	return nil
}

// GetDomainByName fetches the domain list and returns the Domain object
// for the matching domain.
func (c *Client) GetDomainByName(domain string) (*Domain, error) {
	var domains []Domain

	err := c.Request("GET", "domains", nil, &domains)
	if err != nil {
		return nil, err
	}

	for _, d := range domains {
		if !d.Services.DNS {
			// Domains without DNS service cannot have DNS record added
			continue
		}
		if d.Name == domain {
			return &d, nil
		}
	}

	return nil, fmt.Errorf("failed to find matching domain name: %s", domain)
}

// GetDNSRecordByHostData finds the first matching DNS record with the provided host and data.
func (c *Client) GetDNSRecordByHostData(domain Domain, host string, data string) (*DNSRecord, error) {
	var records []DNSRecord

	err := c.Request("GET", fmt.Sprintf("domains/%d/dns", domain.ID), nil, &records)
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

// CreateTXTRecord creates a TXT record wih
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

	return c.Request("POST", fmt.Sprintf("domains/%d/dns", domain.ID), jsonRecord, nil)
}

// DeleteTXTRecord deletes the DNS record matching the provided host and data
func (c *Client) DeleteTXTRecord(domain *Domain, host string, data string) error {

	record, err := c.GetDNSRecordByHostData(*domain, host, data)
	if err != nil {
		return err
	}

	return c.Request("DELETE", fmt.Sprintf("domains/%d/dns/%d", domain.ID, record.ID), nil, nil)
}
