// Package yandex implements a DNS provider for solving the DNS-01
// challenge using yandex DNS.
package yandex

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/xenolf/lego/acme"
)

// YandexAPIURL represents the API endpoint to call.
// TODO: Unexport?
const YandexAPIURL = "https://pddimp.yandex.ru/api2/admin/dns"

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	authEmail string
	authKey   string
}

// NewDNSProvider returns a DNSProvider instance configured for yandex.
// Credentials must be passed in the environment variables: YANDEX_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	email := os.Getenv("YANDEX_EMAIL")
	key := os.Getenv("YANDEX_API_KEY")
	return NewDNSProviderCredentials(email, key)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for yandex.
func NewDNSProviderCredentials(email string, key string) (*DNSProvider, error) {
	if email == "" || key == "" {
		return nil, fmt.Errorf("Yandex credentials missing")
	}

	return &DNSProvider{
		authEmail: email,
		authKey:   key,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (c *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 120 * time.Second, 2 * time.Second
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	parameters := url.Values{}
	parameters.Add("domain", domain)
	parameters.Add("type", "TXT")
	parameters.Add("subdomain", acme.UnFqdn(fqdn))
	parameters.Add("content", value)
	parameters.Add("ttl", "120")

	_, err := c.makeRequest("POST", fmt.Sprintf("/add?%s", parameters.Encode()), nil)
	if err != nil {
		return err
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	record, err := c.findTxtRecord(fqdn, domain)
	if err != nil {
		return err
	}

	parameters := url.Values{}
	parameters.Add("domain", domain)
	parameters.Add("record_id", fmt.Sprintf("%d", record.Id))

	_, err = c.makeRequest("POST", fmt.Sprintf("/del?%s", parameters.Encode()), nil)

	if err != nil {
		return err
	}

	return nil
}

func (c *DNSProvider) findTxtRecord(fqdn string, domain string) (*yandexRecord, error) {
	parameters := url.Values{}
	parameters.Add("domain", domain)
	result, err := c.makeRequest("GET", fmt.Sprintf("/list?%s", parameters.Encode()), nil)
	if err != nil {
		return nil, err
	}

	var records []yandexRecord
	err = json.Unmarshal(result, &records)
	if err != nil {
		return nil, err
	}

	for _, rec := range records {
		foundFqdn := rec.Subdomain + "." + domain
		if foundFqdn == acme.UnFqdn(fqdn) && rec.Type == "TXT" {
			return &rec, nil
		}
	}

	return nil, fmt.Errorf("No existing record found for %s", fqdn)
}

func (c *DNSProvider) makeRequest(method, uri string, body io.Reader) (json.RawMessage, error) {

	// APIResponse represents a response from Yandex API
	type APIResponse struct {
		Domain  string          `json:"domain,omitempty"`
		Success string          `json:"success,omitempty"`
		Error   string          `json:"error,omitempty"`
		Record  json.RawMessage `json:"records,omitempty"`
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", YandexAPIURL, uri), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("admin_email", c.authEmail)
	req.Header.Set("PddToken", c.authKey)

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error querying Yandex API -> %v", err)
	}

	defer resp.Body.Close()

	var r APIResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	if r.Success != "ok" {
		return nil, fmt.Errorf("Yandex API error -> %s", r.Error)
	}

	return r.Record, nil
}

// yandexRecord represents a yandex DNS record
type yandexRecord struct {
	Domain    string `json:"domain"`
	fqdn      string `json:"fqdn"`
	Id        int    `json:"record_id,omitempty"`
	Subdomain string `json:"subdomain,omitempty"`
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	Content   string `json:"content,omitempty"`
	ID        string `json:"id,omitempty"`
	TTL       int    `json:"ttl,omitempty"`
}
