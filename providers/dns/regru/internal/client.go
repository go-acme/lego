package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const defaultBaseURL = "https://api.reg.ru/api/regru2/"

// Client the reg.ru client.
type Client struct {
	username string
	password string

	BaseURL    string
	HTTPClient *http.Client
}

// NewClient Creates a reg.ru client.
func NewClient(username, password string) *Client {
	return &Client{
		username:   username,
		password:   password,
		BaseURL:    defaultBaseURL,
		HTTPClient: http.DefaultClient,
	}
}

// RemoveTxtRecord removes a TXT record.
// https://www.reg.ru/support/help/api2#zone_remove_record
func (c Client) RemoveTxtRecord(domain, subDomain, content string) error {
	request := RemoveRecordRequest{
		Username: c.username,
		Password: c.password,
		Domains: []Domain{
			{DName: domain},
		},
		SubDomain:         subDomain,
		Content:           content,
		RecordType:        "TXT",
		OutputContentType: "plain",
	}

	resp, err := c.do(request, "zone", "remove_record")
	if err != nil {
		return err
	}

	return resp.HasError()
}

// AddTXTRecord adds a TXT record.
// https://www.reg.ru/support/help/api2#zone_add_txt
func (c Client) AddTXTRecord(domain, subDomain, content string) error {
	request := AddTxtRequest{
		Username: c.username,
		Password: c.password,
		Domains: []Domain{
			{DName: domain},
		},
		SubDomain:         subDomain,
		Text:              content,
		OutputContentType: "plain",
	}

	resp, err := c.do(request, "zone", "add_txt")
	if err != nil {
		return err
	}

	return resp.HasError()
}

func (c Client) do(request interface{}, fragments ...string) (*APIResponse, error) {
	endpoint, err := c.createEndpoint(fragments...)
	if err != nil {
		return nil, err
	}

	inputData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Add("input_data", string(inputData))
	query.Add("input_format", "json")
	endpoint.RawQuery = query.Encode()

	resp, err := http.Get(endpoint.String())
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		all, errB := io.ReadAll(resp.Body)
		if errB != nil {
			return nil, fmt.Errorf("API error, status code: %d", resp.StatusCode)
		}

		var apiResp APIResponse
		errB = json.Unmarshal(all, &apiResp)
		if errB != nil {
			return nil, fmt.Errorf("API error, status code: %d, %s", resp.StatusCode, string(all))
		}

		return nil, fmt.Errorf("%w, status code: %d", apiResp, resp.StatusCode)
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	err = json.Unmarshal(all, &apiResp)
	if err != nil {
		return nil, err
	}

	return &apiResp, nil
}

func (c Client) createEndpoint(fragments ...string) (*url.URL, error) {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	return baseURL.JoinPath(fragments...), nil
}
