package zoneee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultEndpoint = "https://api.zone.eu/v2/dns/"

type txtRecord struct {
	// Identifier (identificator)
	ID string `json:"id,omitempty"`
	// Hostname
	Name string `json:"name"`
	// TXT content value
	Destination string `json:"destination"`
	// Can this record be deleted
	Delete bool `json:"delete,omitempty"`
	// Can this record be modified
	Modify bool `json:"modify,omitempty"`
	// API url to get this entity
	ResourceURL string `json:"resource_url,omitempty"`
}

func (d *DNSProvider) addTxtRecord(domain string, record txtRecord) ([]txtRecord, error) {
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(record); err != nil {
		return nil, err
	}

	endpoint := d.config.Endpoint.JoinPath(domain, "txt")

	req, err := http.NewRequest(http.MethodPost, endpoint.String(), reqBody)
	if err != nil {
		return nil, err
	}

	var resp []txtRecord
	if err := d.sendRequest(req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *DNSProvider) getTxtRecords(domain string) ([]txtRecord, error) {
	endpoint := d.config.Endpoint.JoinPath(domain, "txt")

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	var resp []txtRecord
	if err := d.sendRequest(req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *DNSProvider) removeTxtRecord(domain, id string) error {
	endpoint := d.config.Endpoint.JoinPath(domain, "txt", id)

	req, err := http.NewRequest(http.MethodDelete, endpoint.String(), http.NoBody)
	if err != nil {
		return err
	}

	return d.sendRequest(req, nil)
}

func (d *DNSProvider) sendRequest(req *http.Request, result interface{}) error {
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(d.config.Username, d.config.APIKey)

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if err = checkResponse(resp); err != nil {
		return err
	}

	defer resp.Body.Close()

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return fmt.Errorf("unmarshaling %T error [status code=%d]: %w: %s", result, resp.StatusCode, err, string(raw))
	}
	return err
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	if resp.Body == nil {
		return fmt.Errorf("response body is nil, status code=%d", resp.StatusCode)
	}

	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read body: status code=%d, error=%w", resp.StatusCode, err)
	}

	return fmt.Errorf("status code=%d: %s", resp.StatusCode, string(raw))
}
