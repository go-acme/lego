package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	querystring "github.com/google/go-querystring/query"
)

const baseURL = "https://api.internet.bs"

// status SUCCESS, PENDING, FAILURE.
const statusSuccess = "SUCCESS"

// Client is the API client.
type Client struct {
	apiKey   string
	password string

	debug bool

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey, password string) *Client {
	baseURL, _ := url.Parse(baseURL)

	return &Client{
		apiKey:     apiKey,
		password:   password,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// AddRecord The command is intended to add a new DNS record to a specific zone (domain).
func (c *Client) AddRecord(ctx context.Context, query RecordQuery) error {
	var r APIResponse
	err := c.doRequest(ctx, "Add", query, &r)
	if err != nil {
		return err
	}

	if r.Status != statusSuccess {
		return r
	}

	return nil
}

// RemoveRecord The command is intended to remove a DNS record from a specific zone.
func (c *Client) RemoveRecord(ctx context.Context, query RecordQuery) error {
	var r APIResponse
	err := c.doRequest(ctx, "Remove", query, &r)
	if err != nil {
		return err
	}

	if r.Status != statusSuccess {
		return r
	}

	return nil
}

// ListRecords The command is intended to retrieve the list of DNS records for a specific domain.
func (c *Client) ListRecords(ctx context.Context, query ListRecordQuery) ([]Record, error) {
	var l ListResponse
	err := c.doRequest(ctx, "List", query, &l)
	if err != nil {
		return nil, err
	}

	if l.Status != statusSuccess {
		return nil, l.APIResponse
	}

	return l.Records, nil
}

func (c *Client) doRequest(ctx context.Context, action string, params, result any) error {
	endpoint := c.baseURL.JoinPath("Domain", "DnsRecord", action)

	values, err := querystring.Values(params)
	if err != nil {
		return fmt.Errorf("parse query parameters: %w", err)
	}

	values.Set("apiKey", c.apiKey)
	values.Set("password", c.password)
	values.Set("ResponseFormat", "JSON")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	if c.debug {
		return dump(endpoint, resp, result)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func dump(endpoint *url.URL, resp *http.Response, response any) error {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fields := strings.FieldsFunc(endpoint.Path, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	err = os.WriteFile(filepath.Join("fixtures", strings.Join(fields, "_")+".json"), raw, 0o666)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, response)
}
