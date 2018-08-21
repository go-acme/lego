// Package hostingde implements a DNS provider for solving the DNS-01
// challenge using hosting.de.
package hostingde

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// HostingdeAPIURL represents the API endpoint to call.
const HostingdeAPIURL = "https://secure.hosting.de/api/dns/v1/json"

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	authKey  string
	zoneName string
	client   *http.Client
	recordID string
}

// RecordsAddRequest represents a DNS record to add
type RecordsAddRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

// RecordsDeleteRequest represents a DNS record to remove
type RecordsDeleteRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	ID      string `json:"id"`
}

// ZoneConfigObject represents the ZoneConfig-section of a hosting.de API response.
type ZoneConfigObject struct {
	AccountID      string `json:"accountId"`
	EmailAddress   string `json:"emailAddress"`
	ID             string `json:"id"`
	LastChangeDate string `json:"lastChangeDate"`
	MasterIP       string `json:"masterIp"`
	Name           string `json:"name"`
	NameUnicode    string `json:"nameUnicode"`
	SOAValues      struct {
		Expire      int    `json:"expire"`
		NegativeTTL int    `json:"negativeTtl"`
		Refresh     int    `json:"refresh"`
		Retry       int    `json:"retry"`
		Serial      string `json:"serial"`
		TTL         int    `json:"ttl"`
	} `json:"soaValues"`
	Status                string   `json:"status"`
	TemplateValues        string   `json:"templateValues"`
	Type                  string   `json:"type"`
	ZoneTransferWhitelist []string `json:"zoneTransferWhitelist"`
}

// ZoneUpdateError represents an error in a ZoneUpdateResponse
type ZoneUpdateError struct {
	Code          int      `json:"code"`
	ContextObject string   `json:"contextObject"`
	ContextPath   string   `json:"contextPath"`
	Details       []string `json:"details"`
	Text          string   `json:"text"`
	Value         string   `json:"value"`
}

// ZoneUpdateMetadata represents the metadata in a ZoneUpdateResponse
type ZoneUpdateMetadata struct {
	ClientTransactionID string `json:"clientTransactionId"`
	ServerTransactionID string `json:"serverTransactionId"`
}

// ZoneUpdateResponse represents a response from hosting.de API
type ZoneUpdateResponse struct {
	Errors   []ZoneUpdateError  `json:"errors"`
	Metadata ZoneUpdateMetadata `json:"metadata"`
	Warnings []string           `json:"warnings"`
	Status   string             `json:"status"`
	Response struct {
		Records []struct {
			Content          string `json:"content"`
			Type             string `json:"type"`
			ID               string `json:"id"`
			Name             string `json:"name"`
			LastChangeDate   string `json:"lastChangeDate"`
			Priority         int    `json:"priority"`
			RecordTemplateID string `json:"recordTemplateId"`
			ZoneConfigID     string `json:"zoneConfigId"`
			TTL              int    `json:"ttl"`
		} `json:"records"`
		ZoneConfig ZoneConfigObject `json:"zoneConfig"`
	} `json:"response"`
}

// ZoneConfigSelector represents a "minimal" ZoneConfig object used in hosting.de API requests
type ZoneConfigSelector struct {
	Name string `json:"name"`
}

// ZoneUpdateRequest represents a hosting.de API ZoneUpdate request
type ZoneUpdateRequest struct {
	AuthToken          string `json:"authToken"`
	ZoneConfigSelector `json:"zoneConfig"`
	RecordsToAdd       []RecordsAddRequest    `json:"recordsToAdd"`
	RecordsToDelete    []RecordsDeleteRequest `json:"recordsToDelete"`
}

// NewDNSProvider returns a DNSProvider instance configured for hosting.de.
// Credentials must be passed in the environment variables: HOSTINGDE_ZONE_NAME
// and HOSTINGDE_API_KEY
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("HOSTINGDE_API_KEY", "HOSTINGDE_ZONE_NAME")
	if err != nil {
		return nil, fmt.Errorf("Hostingde: %v", err)
	}

	return NewDNSProviderCredentials(values["HOSTINGDE_API_KEY"], values["HOSTINGDE_ZONE_NAME"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for hosting.de.
func NewDNSProviderCredentials(key, zoneName string) (*DNSProvider, error) {
	if key == "" || zoneName == "" {
		return nil, errors.New("Hostingde: API key or Zone Name missing")
	}

	client := &http.Client{Timeout: 30 * time.Second}

	return &DNSProvider{
		authKey:  key,
		zoneName: zoneName,
		client:   client,
		recordID: "",
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 120 * time.Second, 2 * time.Second
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	rec := []RecordsAddRequest{
		RecordsAddRequest{
			Type:    "TXT",
			Name:    acme.UnFqdn(fqdn),
			Content: value,
			TTL:     ttl,
		},
	}

	req := ZoneUpdateRequest{
		AuthToken: d.authKey,
		ZoneConfigSelector: ZoneConfigSelector{
			Name: d.zoneName,
		},
		RecordsToAdd:    rec,
		RecordsToDelete: []RecordsDeleteRequest{},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	var resp ZoneUpdateResponse
	resp, err = d.doRequest(http.MethodPost, "/zoneUpdate", bytes.NewReader(body))

	if err != nil {
		return err
	}

	for _, record := range resp.Response.Records {
		if record.Name == acme.UnFqdn(fqdn) && record.Content == fmt.Sprintf("\"%s\"", value) {
			d.recordID = record.ID
		}
	}

	if d.recordID == "" {
		return fmt.Errorf("error getting ID of just created record")
	}

	return err
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	rec := []RecordsDeleteRequest{
		RecordsDeleteRequest{
			Type:    "TXT",
			Name:    acme.UnFqdn(fqdn),
			Content: value,
			ID:      d.recordID,
		},
	}

	req := ZoneUpdateRequest{
		AuthToken: d.authKey,
		ZoneConfigSelector: ZoneConfigSelector{
			Name: d.zoneName,
		},
		RecordsToAdd:    []RecordsAddRequest{},
		RecordsToDelete: rec,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = d.doRequest(http.MethodPost, "/zoneUpdate", bytes.NewReader(body))
	return err
}

func (d *DNSProvider) doRequest(method, uri string, body io.Reader) (ZoneUpdateResponse, error) {
	var r ZoneUpdateResponse
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", HostingdeAPIURL, uri), body)
	if err != nil {
		return r, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return r, fmt.Errorf("error querying Hostingde API -> %v", err)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return r, err
	}

	if !(r.Status == "success" || r.Status == "pending") {
		strBody := "Unreadable body"
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			strBody = string(body)
		}
		return r, fmt.Errorf("Hostingde API error: the request %s sent the following response: %s", req.URL.String(), strBody)
	}

	return r, nil
}
