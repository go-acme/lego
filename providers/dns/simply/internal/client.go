package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

const defaultBaseURL = "https://api.simply.com/1/"

// Client is a Simply.com API client.
type Client struct {
	HTTPClient  *http.Client
	BaseURL     *url.URL
	accountName string
	apiKey      string
}

// NewClient creates a new Client.
func NewClient(accountName string, apiKey string) (*Client, error) {
	if accountName == "" {
		return nil, errors.New("credentials missing: accountName")
	}

	if apiKey == "" {
		return nil, errors.New("credentials missing: apiKey")
	}

	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		HTTPClient:  &http.Client{},
		BaseURL:     baseURL,
		accountName: accountName,
		apiKey:      apiKey,
	}, nil
}

// apiResponse represents an API response.
type apiResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message"`
	Records json.RawMessage `json:"records,omitempty"`
	Record  json.RawMessage `json:"record,omitempty"`
}

// GetRecords lists all the records in the zone.
func (c *Client) GetRecords(zoneName string) (*[]Record, error) {
	result, err := c.do(zoneName, "/", "GET", nil)
	if err != nil {
		return nil, err
	}

	var rcds []Record
	err = json.Unmarshal(result.Records, &rcds)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response result: %w", err)
	}

	return &rcds, nil
}

// AddRecord adds a record.
func (c *Client) AddRecord(zoneName string, recordBody RecordBody) (*RecordHeader, error) {
	reqBody, err := json.Marshal(recordBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall request body: %w", err)
	}

	result, err := c.do(zoneName, "/", "POST", reqBody)
	if err != nil {
		return nil, err
	}

	var rcd RecordHeader
	err = json.Unmarshal(result.Record, &rcd)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response result: %w", err)
	}

	return &rcd, nil
}

// EditRecord updates a record.
func (c *Client) EditRecord(zoneName string, id int64, recordBody RecordBody) error {
	reqBody, err := json.Marshal(recordBody)
	if err != nil {
		return fmt.Errorf("failed to marshall request body: %w", err)
	}

	_, err = c.do(zoneName, fmt.Sprintf("/%d", id), "PUT", reqBody)
	return err
}

// DeleteRecord deletes a record.
func (c *Client) DeleteRecord(zoneName string, id int64) error {
	_, err := c.do(zoneName, fmt.Sprintf("/%d", id), "DELETE", nil)
	return err
}

func (c *Client) do(zoneName string, endpoint string, reqMethod string, reqBody []byte) (*apiResponse, error) {
	reqURL, err := c.BaseURL.Parse(path.Join(c.BaseURL.Path, fmt.Sprintf("%s/%s/my/products/%s/dns/records", c.accountName, c.apiKey, zoneName), endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	req, err := http.NewRequest(reqMethod, reqURL.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode > 499 {
		return nil, fmt.Errorf("unexpected error: %d", resp.StatusCode)
	}

	response := apiResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Status != 200 {
		return nil, fmt.Errorf("unexpected error: %s", response.Message)
	}

	return &response, nil
}
