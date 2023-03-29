package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const defaultBaseURL = "https://api.beget.com/api/"

// Client the beget.com client.
type Client struct {
	Login  string
	Passwd string

	BaseURL    string
	HTTPClient *http.Client
}

// NewClient Creates a beget.com client.
func NewClient(login, passwd string) *Client {
	return &Client{
		Login:      login,
		Passwd:     passwd,
		BaseURL:    defaultBaseURL,
		HTTPClient: http.DefaultClient,
	}
}

// RemoveTxtRecord removes a TXT record.
// https://beget.com/ru/kb/api/funkczii-upravleniya-dns
func (c Client) RemoveTxtRecord(domain, subDomain string) error {
	request := ChangeRecordsRequest{
		Fqdn:    fmt.Sprintf("%s.%s", subDomain, domain),
		Records: RecordList{},
	}

	resp, err := c.do(request, "dns", "changeRecords")
	if err != nil {
		return err
	}

	return resp.HasError()
}

// AddTXTRecord adds a TXT record.
// https://beget.com/ru/kb/api/funkczii-upravleniya-dns
func (c Client) AddTXTRecord(domain, subDomain, content string) error {
	request := ChangeRecordsRequest{
		Fqdn: fmt.Sprintf("%s.%s", subDomain, domain),
		Records: RecordList{
			TXT: []TxtRecord{
				{Priority: 10, Value: content},
			},
		},
	}

	resp, err := c.do(request, "dns", "changeRecords")
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
	query.Add("login", c.Login)
	query.Add("passwd", c.Passwd)
	query.Add("input_format", "json")
	query.Add("output_format", "json")
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
