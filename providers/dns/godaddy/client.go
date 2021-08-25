package godaddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
)

// DNSRecord a DNS record.
type DNSRecord struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Data     string `json:"data"`
	Priority int    `json:"priority,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
}

func (d *DNSProvider) getRecords(domainZone, rType, recordName string) ([]DNSRecord, error) {
	resource := path.Clean(fmt.Sprintf("/v1/domains/%s/records/%s/%s", domainZone, rType, recordName))

	resp, err := d.makeRequest(http.MethodGet, resource, nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("could not get records: Domain: %s; Record: %s, Status: %v; Body: %s",
			domainZone, recordName, resp.StatusCode, string(bodyBytes))
	}

	var records []DNSRecord
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (d *DNSProvider) updateTxtRecords(records []DNSRecord, domainZone, recordName string) error {
	body, err := json.Marshal(records)
	if err != nil {
		return err
	}

	resource := path.Clean(fmt.Sprintf("/v1/domains/%s/records/TXT/%s", domainZone, recordName))

	var resp *http.Response
	resp, err = d.makeRequest(http.MethodPut, resource, bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
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
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", d.config.APIKey, d.config.APISecret))

	return d.config.HTTPClient.Do(req)
}
