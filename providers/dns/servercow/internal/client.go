package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const baseAPIURL = "https://api.servercow.de/dns/v1/domains"

// Client the Servercow client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client

	username string
	password string
}

// NewClient Creates a Servercow client.
func NewClient(username, password string) *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    baseAPIURL,
		username:   username,
		password:   password,
	}
}

// GetRecords from API.
func (c *Client) GetRecords(domain string) ([]Record, error) {
	req, err := c.createRequest(http.MethodGet, domain, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Note the API always return 200 even if the authentication failed.
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	var records []Record
	err = unmarshal(raw, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// CreateUpdateRecord creates or updates a record.
func (c *Client) CreateUpdateRecord(domain string, data Record) (*Message, error) {
	req, err := c.createRequest(http.MethodPost, domain, &data)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Note the API always return 200 even if the authentication failed.
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	var msg Message
	err = json.Unmarshal(raw, &msg)
	if err != nil {
		return nil, err
	}

	if msg.ErrorMsg != "" {
		return nil, msg
	}

	return &msg, nil
}

// DeleteRecord deletes a record.
func (c *Client) DeleteRecord(domain string, data Record) (*Message, error) {
	req, err := c.createRequest(http.MethodDelete, domain, &data)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Note the API always return 200 even if the authentication failed.
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	var msg Message
	err = json.Unmarshal(raw, &msg)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling %T error: %w: %s", msg, err, string(raw))
	}

	if msg.ErrorMsg != "" {
		return nil, msg
	}

	return &msg, nil
}

func (c *Client) createRequest(method, domain string, payload *Record) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, c.BaseURL+"/"+domain, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Username", c.username)
	req.Header.Set("X-Auth-Password", c.password)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func unmarshal(raw []byte, v interface{}) error {
	err := json.Unmarshal(raw, v)
	if err == nil {
		return nil
	}

	var e *json.UnmarshalTypeError
	if errors.As(err, &e) {
		var apiError Message
		errU := json.Unmarshal(raw, &apiError)
		if errU != nil {
			return fmt.Errorf("unmarshaling %T error: %w: %s", v, err, string(raw))
		}

		return apiError
	}

	return fmt.Errorf("unmarshaling %T error: %w: %s", v, err, string(raw))
}
