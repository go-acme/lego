package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.simply.com/1/"

// Client is a Simply.com API client.
type Client struct {
	HTTPClient  *http.Client
	baseURL     *url.URL
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
		HTTPClient:  &http.Client{Timeout: 5 * time.Second},
		baseURL:     baseURL,
		accountName: accountName,
		apiKey:      apiKey,
	}, nil
}

// GetRecords lists all the records in the zone.
func (c *Client) GetRecords(zoneName string) ([]Record, error) {
	resp, err := c.do(zoneName, "/", http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	var records []Record
	err = json.Unmarshal(resp.Records, &records)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response result: %w", err)
	}

	return records, nil
}

// AddRecord adds a record.
func (c *Client) AddRecord(zoneName string, record Record) (int64, error) {
	reqBody, err := json.Marshal(record)
	if err != nil {
		return 0, fmt.Errorf("failed to marshall request body: %w", err)
	}

	resp, err := c.do(zoneName, "/", http.MethodPost, reqBody)
	if err != nil {
		return 0, err
	}

	var rcd recordHeader
	err = json.Unmarshal(resp.Record, &rcd)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal response result: %w", err)
	}

	return rcd.ID, nil
}

// EditRecord updates a record.
func (c *Client) EditRecord(zoneName string, id int64, record Record) error {
	reqBody, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshall request body: %w", err)
	}

	_, err = c.do(zoneName, fmt.Sprintf("%d", id), http.MethodPut, reqBody)
	return err
}

// DeleteRecord deletes a record.
func (c *Client) DeleteRecord(zoneName string, id int64) error {
	_, err := c.do(zoneName, fmt.Sprintf("%d", id), http.MethodDelete, nil)
	return err
}

func (c *Client) do(zoneName string, endpoint string, reqMethod string, reqBody []byte) (*apiResponse, error) {
	reqURL := c.baseURL.JoinPath(c.accountName, c.apiKey, "my", "products", zoneName, "dns", "records", endpoint)

	req, err := http.NewRequest(reqMethod, strings.TrimSuffix(reqURL.String(), "/"), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, fmt.Errorf("unexpected error: %d", resp.StatusCode)
	}

	response := apiResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Status != http.StatusOK {
		return nil, fmt.Errorf("unexpected error: %s", response.Message)
	}

	return &response, nil
}
