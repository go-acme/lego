package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

const (
	defaultBaseURL = "https://api.dns.constellix.com"
	defaultVersion = "v1"
)

// Client the Constellix client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for communicating with the API
	Domains    *DomainService
	TxtRecords *TxtRecordService
}

// NewClient Creates a Constellix client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	client := &Client{
		BaseURL:    defaultBaseURL,
		HTTPClient: httpClient,
	}

	client.common.client = client
	client.Domains = (*DomainService)(&client.common)
	client.TxtRecords = (*TxtRecordService)(&client.common)

	return client
}

type service struct {
	client *Client
}

// do sends an API request and returns the API response.
func (c *Client) do(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	err = checkResponse(resp)
	if err != nil {
		return err
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	if err = json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("unmarshaling %T error: %w: %s", v, err, string(raw))
	}

	return nil
}

func (c *Client) createEndpoint(fragment ...string) (string, error) {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}

	endpoint, err := baseURL.Parse(path.Join(fragment...))
	if err != nil {
		return "", err
	}

	return endpoint.String(), nil
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err == nil && data != nil {
		msg := &APIError{StatusCode: resp.StatusCode}

		if json.Unmarshal(data, msg) != nil {
			return fmt.Errorf("API error: status code: %d: %v", resp.StatusCode, string(data))
		}

		switch resp.StatusCode {
		case http.StatusNotFound:
			return &NotFound{APIError: msg}
		case http.StatusBadRequest:
			return &BadRequest{APIError: msg}
		default:
			return msg
		}
	}

	return fmt.Errorf("API error, status code: %d", resp.StatusCode)
}
