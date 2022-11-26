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

const defaultBaseURL = "https://dns-service.iran.liara.ir"

// Client a Liara DNS API client.
type Client struct {
	apiKey     string
	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
	}
}

// GetRecords gets the records of a domain.
// https://dns-service.iran.liara.ir/swagger
func (c Client) GetRecords(domainName string) ([]Record, error) {
	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "api", "v1", "zones", domainName, "dns-records"))
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	var response RecordsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// CreateRecord creates a record.
func (c Client) CreateRecord(domainName string, record Record) (*Record, error) {
	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "api", "v1", "zones", domainName, "dns-records"))
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshal request data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return nil, readError(resp)
	}

	var response RecordResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// GetRecord gets a specific record.
func (c Client) GetRecord(domainName, recordID string) (*Record, error) {
	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "api", "v1", "zones", domainName, "dns-records", recordID))
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	var response RecordResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// DeleteRecord deletes a record.
func (c Client) DeleteRecord(domainName, recordID string) error {
	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "api", "v1", "zones", domainName, "dns-records", recordID))
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return readError(resp)
	}

	return nil
}

func readError(resp *http.Response) error {
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("API error (status code: %d)", resp.StatusCode)
	}

	var apiError APIError
	err = json.Unmarshal(all, &apiError)
	if err != nil {
		return fmt.Errorf("API error (status code: %d): %s", resp.StatusCode, string(all))
	}

	return fmt.Errorf("API error (status code: %d): %w", resp.StatusCode, &apiError)
}
