package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://pddimp.yandex.ru/api2/admin/dns"

const successCode = "ok"

const pddTokenHeader = "PddToken"

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	pddToken   string
}

func NewClient(pddToken string) (*Client, error) {
	if pddToken == "" {
		return nil, errors.New("PDD token is required")
	}
	return &Client{
		HTTPClient: &http.Client{},
		BaseURL:    defaultBaseURL,
		pddToken:   pddToken,
	}, nil
}

func (c *Client) AddRecord(data Record) (*Record, error) {
	resp, err := c.postForm("/add", data)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API response error: %d", resp.StatusCode)
	}

	r := AddResponse{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	if r.Success != successCode {
		return nil, fmt.Errorf("error during record addition: %s", r.Error)
	}

	return r.Record, nil
}

func (c *Client) RemoveRecord(data Record) (int, error) {
	resp, err := c.postForm("/del", data)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	r := RemoveResponse{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return 0, err
	}

	if r.Success != successCode {
		return 0, fmt.Errorf("error during record addition: %s", r.Error)
	}

	return r.RecordID, nil
}

func (c *Client) GetRecords(domain string) ([]Record, error) {
	resp, err := c.get("/list", struct {
		Domain string `url:"domain"`
	}{Domain: domain})
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	r := ListResponse{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	if r.Success != successCode {
		return nil, fmt.Errorf("error during record addition: %s", r.Error)
	}

	return r.Records, nil
}

func (c *Client) postForm(uri string, data interface{}) (*http.Response, error) {
	values, err := query.Values(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+uri, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(pddTokenHeader, c.pddToken)

	return c.HTTPClient.Do(req)
}

func (c *Client) get(uri string, data interface{}) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(pddTokenHeader, c.pddToken)

	values, err := query.Values(data)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = values.Encode()

	return c.HTTPClient.Do(req)
}
