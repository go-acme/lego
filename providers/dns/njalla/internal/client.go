package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const apiEndpoint = "https://njal.la/api/1/"

// Client is a Njalla API client.
type Client struct {
	HTTPClient  *http.Client
	apiEndpoint string
	token       string
}

// NewClient creates a new Client.
func NewClient(token string) *Client {
	return &Client{
		HTTPClient:  &http.Client{Timeout: 5 * time.Second},
		apiEndpoint: apiEndpoint,
		token:       token,
	}
}

// AddRecord adds a record.
func (c *Client) AddRecord(record Record) (*Record, error) {
	data := APIRequest{
		Method: "add-record",
		Params: record,
	}

	result, err := c.do(data)
	if err != nil {
		return nil, err
	}

	var rcd Record
	err = json.Unmarshal(result, &rcd)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response result: %w", err)
	}

	return &rcd, nil
}

// RemoveRecord removes a record.
func (c *Client) RemoveRecord(id string, domain string) error {
	data := APIRequest{
		Method: "remove-record",
		Params: Record{
			ID:     id,
			Domain: domain,
		},
	}

	_, err := c.do(data)
	if err != nil {
		return err
	}

	return nil
}

// ListRecords list the records for one domain.
func (c *Client) ListRecords(domain string) ([]Record, error) {
	data := APIRequest{
		Method: "list-records",
		Params: Record{
			Domain: domain,
		},
	}

	result, err := c.do(data)
	if err != nil {
		return nil, err
	}

	var rcds Records
	err = json.Unmarshal(result, &rcds)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response result: %w", err)
	}

	return rcds.Records, nil
}

func (c *Client) do(data APIRequest) (json.RawMessage, error) {
	req, err := c.createRequest(data)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected error: %d", resp.StatusCode)
	}

	apiResponse := APIResponse{}
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResponse.Error != nil {
		return nil, apiResponse.Error
	}

	return apiResponse.Result, nil
}

func (c *Client) createRequest(data APIRequest) (*http.Request, error) {
	reqBody, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.apiEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Njalla "+c.token)

	return req, nil
}
