package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	defaultEndpoint  = "https://api.scaleway.com/domain/v2alpha2"
	uriUpdateRecords = "/dns-zones/%s/records"
	operationSet     = "set"
	operationDelete  = "delete"
	operationAdd     = "add"
)

// APIError represents an error response from the API.
type APIError struct {
	Message string `json:"message"`
}

func (a APIError) Error() string {
	return a.Message
}

// Record represents a DNS record
type Record struct {
	Data     string `json:"data,omitempty"`
	Name     string `json:"name,omitempty"`
	Priority uint32 `json:"priority,omitempty"`
	TTL      uint32 `json:"ttl,omitempty"`
	Type     string `json:"type,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

// RecordChangeAdd represents a list of add operations.
type RecordChangeAdd struct {
	Records []*Record `json:"records,omitempty"`
}

// RecordChangeSet represents a list of set operations.
type RecordChangeSet struct {
	Data    string    `json:"data,omitempty"`
	Name    string    `json:"name,omitempty"`
	TTL     uint32    `json:"ttl,omitempty"`
	Type    string    `json:"type,omitempty"`
	Records []*Record `json:"records,omitempty"`
}

// RecordChangeDelete represents a list of delete operations.
type RecordChangeDelete struct {
	Data string `json:"data,omitempty"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

// UpdateDNSZoneRecordsRequest represents a request to update DNS records on the API.
type UpdateDNSZoneRecordsRequest struct {
	DNSZone          string        `json:"dns_zone,omitempty"`
	Changes          []interface{} `json:"changes,omitempty"`
	ReturnAllRecords bool          `json:"return_all_records,omitempty"`
}

// ClientOpts represents options to init client.
type ClientOpts struct {
	BaseURL string
	Token   string
}

// Client represents DNS client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient returns a client instance.
func NewClient(opts ClientOpts, httpClient *http.Client) *Client {
	baseURL := defaultEndpoint
	if opts.BaseURL != "" {
		baseURL = opts.BaseURL
	}

	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &Client{
		token:      opts.Token,
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// AddRecord adds Record for given zone.
func (c *Client) AddRecord(zone string, record Record) error {
	changes := map[string]RecordChangeAdd{
		operationAdd: {
			Records: []*Record{&record},
		},
	}

	request := UpdateDNSZoneRecordsRequest{
		DNSZone:          zone,
		Changes:          []interface{}{changes},
		ReturnAllRecords: false,
	}

	uri := fmt.Sprintf(uriUpdateRecords, zone)
	req, err := c.newRequest(http.MethodPatch, uri, request)
	if err != nil {
		return err
	}

	return c.do(req)
}

// SetRecord sets a unique Record for given zone.
func (c *Client) SetRecord(zone string, record Record) error {
	changes := map[string]RecordChangeSet{
		operationSet: {
			Name:    record.Name,
			Type:    record.Type,
			Records: []*Record{&record},
		},
	}

	request := UpdateDNSZoneRecordsRequest{
		DNSZone:          zone,
		Changes:          []interface{}{changes},
		ReturnAllRecords: false,
	}

	uri := fmt.Sprintf(uriUpdateRecords, zone)
	req, err := c.newRequest(http.MethodPatch, uri, request)
	if err != nil {
		return err
	}

	return c.do(req)
}

// DeleteRecord deletes a Record for given zone.
func (c *Client) DeleteRecord(zone string, record Record) error {
	delRecord := map[string]RecordChangeDelete{
		operationDelete: {
			Name: record.Name,
			Type: record.Type,
			Data: record.Data,
		},
	}

	request := UpdateDNSZoneRecordsRequest{
		DNSZone:          zone,
		Changes:          []interface{}{delRecord},
		ReturnAllRecords: false,
	}

	uri := fmt.Sprintf(uriUpdateRecords, zone)
	req, err := c.newRequest(http.MethodPatch, uri, request)
	if err != nil {
		return err
	}

	return c.do(req)
}

func (c *Client) newRequest(method, uri string, body interface{}) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request body with error: %w", err)
		}
	}

	req, err := http.NewRequest(method, c.baseURL+uri, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create new http request with error: %w", err)
	}

	req.Header.Add("X-auth-token", c.token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed with error: %w", err)
	}

	err = checkResponse(resp)
	if err != nil {
		return err
	}

	return checkResponse(resp)
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= http.StatusBadRequest || resp.StatusCode < http.StatusOK {
		if resp.Body == nil {
			return fmt.Errorf("request failed with status code %d and empty body", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		apiError := APIError{}
		err = json.Unmarshal(body, &apiError)
		if err != nil {
			return fmt.Errorf("request failed with status code %d, response body: %s", resp.StatusCode, string(body))
		}

		return fmt.Errorf("request failed with status code %d: %w", resp.StatusCode, apiError)
	}

	return nil
}
