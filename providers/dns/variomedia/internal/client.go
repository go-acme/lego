package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://api.variomedia.de"

type Client struct {
	apiToken   string
	baseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(apiToken string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiToken:   apiToken,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c Client) CreateDNSRecord(record DNSRecord) (*CreateDNSRecordResponse, error) {
	endpoint := c.baseURL.JoinPath("dns-records")

	data := CreateDNSRecordRequest{Data: Data{
		Type:       "dns-record",
		Attributes: record,
	}}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var result CreateDNSRecordResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c Client) DeleteDNSRecord(id string) (*DeleteRecordResponse, error) {
	endpoint := c.baseURL.JoinPath("dns-records", id)

	req, err := http.NewRequest(http.MethodDelete, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	var result DeleteRecordResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c Client) GetJob(id string) (*GetJobResponse, error) {
	endpoint := c.baseURL.JoinPath("queue-jobs", id)

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	var result GetJobResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c Client) do(req *http.Request, data interface{}) error {
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/vnd.variomedia.v1+json")
	req.Header.Set("Authorization", "token "+c.apiToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		all, _ := io.ReadAll(resp.Body)

		var e APIError
		err = json.Unmarshal(all, &e)
		if err != nil {
			return fmt.Errorf("%d: %s", resp.StatusCode, string(all))
		}

		return e
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, data)
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(content))
	}

	return nil
}
