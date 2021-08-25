package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://public-api.sonic.net/dyndns"

type APIResponse struct {
	Message string `json:"message"`
	Result  int    `json:"result"`
}

// Record holds the Sonic API representation of a Domain Record.
type Record struct {
	UserID   string `json:"userid"`
	APIKey   string `json:"apikey"`
	Hostname string `json:"hostname"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	Type     string `json:"type"`
}

// Client Sonic client.
type Client struct {
	userID     string
	apiKey     string
	baseURL    string
	HTTPClient *http.Client
}

// NewClient creates a Client.
func NewClient(userID, apiKey string) (*Client, error) {
	if userID == "" || apiKey == "" {
		return nil, errors.New("credentials are missing")
	}

	return &Client{
		userID:     userID,
		apiKey:     apiKey,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// SetRecord creates or updates a TXT records.
// Sonic does not provide a delete record API endpoint.
// https://public-api.sonic.net/dyndns#updating_or_adding_host_records
func (c *Client) SetRecord(hostname string, value string, ttl int) error {
	payload := &Record{
		UserID:   c.userID,
		APIKey:   c.apiKey,
		Hostname: hostname,
		Value:    value,
		TTL:      ttl,
		Type:     "TXT",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, c.baseURL+"/host", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("content-type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	r := APIResponse{}
	err = json.Unmarshal(raw, &r)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w: %s", err, string(raw))
	}

	if r.Result != 200 {
		return fmt.Errorf("API response code: %d, %s", r.Result, r.Message)
	}

	return nil
}
