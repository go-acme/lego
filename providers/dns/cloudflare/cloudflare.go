// Package cloudflare implements a DNS provider for solving the DNS-01
// challenge using cloudflare DNS.
package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
)

// CloudFlareAPIURL represents the API endpoint to call.
// TODO: Unexport?
const CloudFlareAPIURL = "https://api.cloudflare.com/client/v4"

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	authEmail string
	authKey   string
}

// NewDNSProvider returns a DNSProvider instance configured for cloudflare.
// Credentials must be passed in the environment variables: CLOUDFLARE_EMAIL
// and CLOUDFLARE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	email := os.Getenv("CLOUDFLARE_EMAIL")
	key := os.Getenv("CLOUDFLARE_API_KEY")
	return NewDNSProviderCredentials(email, key)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for cloudflare.
func NewDNSProviderCredentials(email, key string) (*DNSProvider, error) {
	if email == "" || key == "" {
		return nil, fmt.Errorf("CloudFlare credentials missing")
	}

	return &DNSProvider{
		authEmail: email,
		authKey:   key,
	}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (c *DNSProvider) Present(domain *acme.Domain, token, keyAuth string) error {
	fqdn, value, _ := domain.GetDNS01Record(keyAuth)
	zoneID, err := c.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	rec := cloudFlareRecord{
		Type:    "TXT",
		Name:    domain.Domain,
		Content: value,
		TTL:     120,
	}

	body, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	_, err = c.makeRequest("POST", fmt.Sprintf("/zones/%s/dns_records", zoneID), bytes.NewReader(body))
	if err != nil {
		return err
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (c *DNSProvider) CleanUp(domain *acme.Domain, token, keyAuth string) error {
	record, err := c.findTxtRecord(domain)
	if err != nil {
		return err
	}

	_, err = c.makeRequest("DELETE", fmt.Sprintf("/zones/%s/dns_records/%s", record.ZoneID, record.ID), nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *DNSProvider) getHostedZoneID(fqdn string) (string, error) {
	// HostedZone represents a CloudFlare DNS zone
	type HostedZone struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	result, err := c.makeRequest("GET", "/zones?per_page=1000", nil)
	if err != nil {
		return "", err
	}

	var zones []HostedZone
	err = json.Unmarshal(result, &zones)
	if err != nil {
		return "", err
	}

	var hostedZone HostedZone
	for _, zone := range zones {
		domain := acme.NewDomain(zone.Name)
		name := domain.GetFqdn()
		if strings.HasSuffix(fqdn, name) {
			if len(zone.Name) > len(hostedZone.Name) {
				hostedZone = zone
			}
		}
	}
	if hostedZone.ID == "" {
		return "", fmt.Errorf("No matching CloudFlare zone found for %s", fqdn)
	}

	return hostedZone.ID, nil
}

func (c *DNSProvider) findTxtRecord(domain *acme.Domain) (*cloudFlareRecord, error) {
	zoneID, err := c.getHostedZoneID(domain.GetFqdn())
	if err != nil {
		return nil, err
	}

	result, err := c.makeRequest("GET", fmt.Sprintf("/zones/%s/dns_records?per_page=1000", zoneID), nil)
	if err != nil {
		return nil, err
	}

	var records []cloudFlareRecord
	err = json.Unmarshal(result, &records)
	if err != nil {
		return nil, err
	}

	for _, rec := range records {
		if rec.Name == domain.Domain && rec.Type == "TXT" {
			return &rec, nil
		}
	}

	return nil, fmt.Errorf("No existing record found for %s", domain.GetFqdn())
}

func (c *DNSProvider) makeRequest(method, uri string, body io.Reader) (json.RawMessage, error) {
	// APIError contains error details for failed requests
	type APIError struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	}

	// APIResponse represents a response from CloudFlare API
	type APIResponse struct {
		Success bool            `json:"success"`
		Errors  []*APIError     `json:"errors"`
		Result  json.RawMessage `json:"result"`
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", CloudFlareAPIURL, uri), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Email", c.authEmail)
	req.Header.Set("X-Auth-Key", c.authKey)
	//req.Header.Set("User-Agent", userAgent())

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error querying API -> %v", err)
	}

	defer resp.Body.Close()

	var r APIResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	if !r.Success {
		if len(r.Errors) > 0 {
			return nil, fmt.Errorf("API error -> %d: %s", r.Errors[0].Code, r.Errors[0].Message)
		}
		return nil, fmt.Errorf("API error")
	}

	return r.Result, nil
}

// cloudFlareRecord represents a CloudFlare DNS record
type cloudFlareRecord struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	ID      string `json:"id,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
	ZoneID  string `json:"zone_id,omitempty"`
}
