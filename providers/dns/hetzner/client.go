package hetzner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
)

// DNSRecord a DNS record
type DNSRecord struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value"`
	Priority int    `json:"priority,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	ID       string
	ZoneID   string `json:"zone_id,omitempty"`
}

type Records struct {
	Records []DNSRecord
}

type Zone struct {
	ID   string
	Name string
}

type Zones struct {
	Zones []Zone
}

func (d *DNSProvider) getTxtRecord(name, value, zoneID string) (*DNSRecord, error) {
	resource := path.Clean(fmt.Sprintf("/api/v1/records?zone_id=%s", zoneID))

	resp, err := d.makeRequest(http.MethodGet, resource, nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("could not get records: zone ID: %s; Record: %s, Status: %v; Body: %s",
			zoneID, name, resp.StatusCode, string(bodyBytes))
	}

	var records Records
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		return nil, err
	}

	for _, record := range records.Records {
		if record.Name == name && record.Value == value {
			return &record, nil
		}
	}

	return nil, fmt.Errorf("could not find record: zone ID: %s; Record: %s", zoneID, name)
}

func (d *DNSProvider) deleteTxtRecord(domain string, record DNSRecord) error {
	resource := path.Clean(fmt.Sprintf("/api/v1/records/%s", record.ID))

	var resp *http.Response
	resp, err := d.makeRequest(http.MethodDelete, resource, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not delete record Domain: %s; %v; Status: %v", domain, record.Name, resp.StatusCode)
	}

	return nil
}

func (d *DNSProvider) createTxtRecord(record DNSRecord) error {
	body, err := json.Marshal(record)
	if err != nil {
		return err
	}

	resource := path.Clean("/api/v1/records")

	var resp *http.Response
	resp, err = d.makeRequest(http.MethodPost, resource, bytes.NewReader(body))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("could not create record %v; Status: %v; Body: %s", string(body), resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (d *DNSProvider) makeRequest(method, uri string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", defaultBaseURL, uri), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-API-Token", d.config.APIKey)

	return d.config.HTTPClient.Do(req)
}

func (d *DNSProvider) getZoneID(domain string) (zoneID string, err error) {
	resource := path.Clean("/api/v1/zones")

	var resp *http.Response
	resp, err = d.makeRequest(http.MethodGet, resource, nil)
	if err != nil {
		return "", fmt.Errorf("could not get zones Domain %v: %v", domain, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("could not get zones Domain %v; Status: %v", domain, resp.StatusCode)
	}

	var zones Zones
	err = json.NewDecoder(resp.Body).Decode(&zones)
	if err != nil {
		return "", err
	}

	for _, zone := range zones.Zones {
		if zone.Name == domain {
			return zone.ID, nil
		}
	}

	return "", fmt.Errorf("could not get zones Domain %v: zone for domain%v not found", domain, domain)
}
