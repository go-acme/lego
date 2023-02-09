package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

// DefaultBaseURL represents the API endpoint to call.
const DefaultBaseURL = "https://api.godaddy.com"

type Client struct {
	HTTPClient *http.Client
	baseURL    *url.URL
	apiKey     string
	apiSecret  string
}

func NewClient(apiKey string, apiSecret string) *Client {
	baseURL, _ := url.Parse(DefaultBaseURL)

	return &Client{
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    baseURL,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
	}
}

func (d *Client) GetRecords(domainZone, rType, recordName string) ([]DNSRecord, error) {
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

func (d *Client) UpdateTxtRecords(records []DNSRecord, domainZone, recordName string) error {
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

func (d *Client) makeRequest(method, uri string, body io.Reader) (*http.Response, error) {
	endpoint := d.baseURL.JoinPath(uri)

	req, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", d.apiKey, d.apiSecret))

	return d.HTTPClient.Do(req)
}
