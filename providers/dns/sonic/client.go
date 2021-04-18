package sonic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Client Sonic client.
type Client struct {
	userId     string
	apiKey  string
	HTTPClient *http.Client
}

// Record holds the Sonic API representation of a Domain Record.
type Record struct {
	UserId     string `json:"userid"`
	ApiKey     string `json:"apikey"`
	Hostname   string `json:"hostname"`
	Value      string  `json:"value"`
	TTL        int     `json:"ttl"`
	Type       string  `json:"type"`
}


// NewClient creates a sonic client based on DNSMadeEasy's LEGO library
func NewClient(userId, apiKey string) (*Client, error) {
	if userId == "" {
		return nil, errors.New("credentials missing: userId created via https://public-api.sonic.net/dyndns#requesting_an_api_key")
	}

	if apiKey == "" {
		return nil, errors.New("credentials missing: apiKey created via https://public-api.sonic.net/dyndns#requesting_an_api_key")
	}

	return &Client{
		userId:     userId,
		apiKey:  apiKey,
		HTTPClient: &http.Client{},
	}, nil
}

// CreateOrUpdateRecord creates or updates a TXT records.
// Sonic does not provide a delete record API service
// Example CURL from https://public-api.sonic.net/dyndns#updating_or_adding_host_records
// # curl -X PUT -H "Content-Type: application/json" --data '{"userid":"12345","apikey":"4d6fbf2f9ab0fa11697470918d37625851fc0c51","hostname":"foo.example.com","value":"209.204.190.64","type":"A"}' https://public-api.sonic.net/dyndns/host
func (c *Client) CreateOrUpdateRecord(hostname string, value string, ttl int) error {
	resp, err := c.sendRequest(http.MethodPut, &Record{UserId: c.userId, ApiKey:c.apiKey, Hostname:hostname, Value:value, Type: "TXT", TTL: ttl})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) sendRequest(method string, payload interface{}) (*http.Response, error) {
	url := "https://public-api.sonic.net/dyndns/host"

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("request failed with HTTP status code %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("request failed with HTTP status code %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}
