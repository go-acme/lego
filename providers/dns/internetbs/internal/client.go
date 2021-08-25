package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	querystring "github.com/google/go-querystring/query"
)

const baseURL = "https://api.internet.bs"

// status SUCCESS, PENDING, FAILURE.
const statusSuccess = "SUCCESS"

// Client is the API client.
type Client struct {
	HTTPClient *http.Client
	baseURL    *url.URL
	debug      bool

	apiKey   string
	password string
}

// NewClient creates a new Client.
func NewClient(apiKey string, password string) *Client {
	baseURL, _ := url.Parse(baseURL)

	return &Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		apiKey:     apiKey,
		password:   password,
	}
}

// AddRecord The command is intended to add a new DNS record to a specific zone (domain).
func (c Client) AddRecord(query RecordQuery) error {
	var r APIResponse
	err := c.do("Add", query, &r)
	if err != nil {
		return err
	}

	if r.Status != statusSuccess {
		return r
	}

	return nil
}

// RemoveRecord The command is intended to remove a DNS record from a specific zone.
func (c Client) RemoveRecord(query RecordQuery) error {
	var r APIResponse
	err := c.do("Remove", query, &r)
	if err != nil {
		return err
	}

	if r.Status != statusSuccess {
		return r
	}

	return nil
}

// ListRecords The command is intended to retrieve the list of DNS records for a specific domain.
func (c Client) ListRecords(query ListRecordQuery) ([]Record, error) {
	var l ListResponse
	err := c.do("List", query, &l)
	if err != nil {
		return nil, err
	}

	if l.Status != statusSuccess {
		return nil, l.APIResponse
	}

	return l.Records, nil
}

func (c Client) do(action string, params interface{}, response interface{}) error {
	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "Domain", "DnsRecord", action))
	if err != nil {
		return fmt.Errorf("create endpoint: %w", err)
	}

	values, err := querystring.Values(params)
	if err != nil {
		return fmt.Errorf("parse query parameters: %w", err)
	}

	values.Set("apiKey", c.apiKey)
	values.Set("password", c.password)
	values.Set("ResponseFormat", "JSON")

	resp, err := c.HTTPClient.PostForm(endpoint.String(), values)
	if err != nil {
		return fmt.Errorf("post request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status code: %d, %s", resp.StatusCode, string(data))
	}

	if c.debug {
		return dump(endpoint, resp, response)
	}

	return json.NewDecoder(resp.Body).Decode(response)
}

func dump(endpoint *url.URL, resp *http.Response, response interface{}) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fields := strings.FieldsFunc(endpoint.Path, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	err = os.WriteFile(filepath.Join("fixtures", strings.Join(fields, "_")+".json"), data, 0o666)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, response)
}
