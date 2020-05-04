package luadns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type errorResponse struct {
	Status    string `json:"status"`
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
}

type DNSZone struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Synced         bool   `json:"synced"`
	QueriesCount   int    `json:"queries_count"`
	RecordsCount   int    `json:"records_count"`
	AliasesCount   int    `json:"aliases_count"`
	RedirectsCount int    `json:"redirects_count"`
	ForwardsCount  int    `json:"forwards_count"`
	TemplateID     int    `json:"template_id"`
}

type NewDNSRecord struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

type DNSRecord struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	ZoneID  int    `json:"zone_id"`
}

func (d *DNSProvider) listZones(domainZone string) ([]DNSZone, error) {
	resource := "/v1/zones"

	resp, err := d.makeRequest(http.MethodGet, resource, nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		var errResp errorResponse
		err = json.Unmarshal(bodyBytes, &errResp)
		if err == nil {
			return nil, fmt.Errorf("could not get zones: Domain: %s; Status: %v %s; Message: %s",
				domainZone, resp.StatusCode, errResp.Status, errResp.Message)
		}
		return nil, fmt.Errorf("could not get zones: Domain: %s; Status: %v; Body: %s",
			domainZone, resp.StatusCode, string(bodyBytes))
	}

	var zones []DNSZone
	err = json.NewDecoder(resp.Body).Decode(&zones)
	if err != nil {
		return nil, err
	}

	return zones, nil
}

func (d *DNSProvider) createRecord(zone DNSZone, newRecord NewDNSRecord) (*DNSRecord, error) {
	body, err := json.Marshal(newRecord)
	if err != nil {
		return nil, err
	}

	resource := fmt.Sprintf("/v1/zones/%d/records", zone.ID)

	var resp *http.Response
	resp, err = d.makeRequest(http.MethodPost, resource, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		var errResp errorResponse
		err = json.Unmarshal(bodyBytes, &errResp)
		if err == nil {
			return nil, fmt.Errorf("could not create record %v; Status: %v %s; Message: %s",
				string(body), resp.StatusCode, errResp.Status, errResp.Message)
		}
		return nil, fmt.Errorf("could not create record %v; Status: %v; Body: %s",
			string(body), resp.StatusCode, string(bodyBytes))
	}

	var record *DNSRecord
	err = json.NewDecoder(resp.Body).Decode(&record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (d *DNSProvider) deleteRecord(record *DNSRecord) error {
	body, err := json.Marshal(record)
	if err != nil {
		return err
	}

	resource := fmt.Sprintf("/v1/zones/%d/records/%d", record.ZoneID, record.ID)

	var resp *http.Response
	resp, err = d.makeRequest(http.MethodDelete, resource, bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		var errResp errorResponse
		err = json.Unmarshal(bodyBytes, &errResp)
		if err == nil {
			return fmt.Errorf("could not delete record %v; Status: %v %s; Message: %s",
				string(body), resp.StatusCode, errResp.Status, errResp.Message)
		}
		return fmt.Errorf("could not delete record %v; Status: %v; Body: %s",
			string(body), resp.StatusCode, string(bodyBytes))
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
	req.SetBasicAuth(d.config.APIUsername, d.config.APIToken)

	return d.config.HTTPClient.Do(req)
}
