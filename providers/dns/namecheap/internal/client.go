package internal

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// Default API endpoints.
const (
	DefaultBaseURL = "https://api.namecheap.com/xml.response"
	SandboxBaseURL = "https://api.sandbox.namecheap.com/xml.response"
)

// Client the API client for Namecheap.
type Client struct {
	apiUser  string
	apiKey   string
	clientIP string

	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiUser, apiKey, clientIP string) *Client {
	return &Client{
		apiUser:    apiUser,
		apiKey:     apiKey,
		clientIP:   clientIP,
		BaseURL:    DefaultBaseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetHosts reads the full list of DNS host records.
// https://www.namecheap.com/support/api/methods/domains-dns/get-hosts.aspx
func (c *Client) GetHosts(ctx context.Context, sld, tld string) ([]Record, error) {
	request, err := c.newRequestGet(ctx, "namecheap.domains.dns.getHosts",
		addParam("SLD", sld),
		addParam("TLD", tld),
	)
	if err != nil {
		return nil, err
	}

	var ghr getHostsResponse
	err = c.do(request, &ghr)
	if err != nil {
		return nil, err
	}

	if len(ghr.Errors) > 0 {
		return nil, ghr.Errors[0]
	}

	return ghr.Hosts, nil
}

// SetHosts writes the full list of DNS host records .
// https://www.namecheap.com/support/api/methods/domains-dns/set-hosts.aspx
func (c *Client) SetHosts(ctx context.Context, sld, tld string, hosts []Record) error {
	req, err := c.newRequestPost(ctx, "namecheap.domains.dns.setHosts",
		addParam("SLD", sld),
		addParam("TLD", tld),
		func(values url.Values) {
			for i, h := range hosts {
				ind := strconv.Itoa(i + 1)
				values.Add("HostName"+ind, h.Name)
				values.Add("RecordType"+ind, h.Type)
				values.Add("Address"+ind, h.Address)
				values.Add("MXPref"+ind, h.MXPref)
				values.Add("TTL"+ind, h.TTL)
			}
		},
	)
	if err != nil {
		return err
	}

	var shr setHostsResponse
	err = c.do(req, &shr)
	if err != nil {
		return err
	}

	if len(shr.Errors) > 0 {
		return shr.Errors[0]
	}
	if shr.Result.IsSuccess != "true" {
		return errors.New("setHosts failed")
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	return xml.Unmarshal(raw, result)
}

func (c *Client) newRequestGet(ctx context.Context, cmd string, params ...func(url.Values)) (*http.Request, error) {
	query := c.makeQuery(cmd, params...)

	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	return req, nil
}

func (c *Client) newRequestPost(ctx context.Context, cmd string, params ...func(url.Values)) (*http.Request, error) {
	query := c.makeQuery(cmd, params...)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, strings.NewReader(query.Encode()))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func (c *Client) makeQuery(cmd string, params ...func(url.Values)) url.Values {
	queryParams := make(url.Values)
	queryParams.Set("ApiUser", c.apiUser)
	queryParams.Set("ApiKey", c.apiKey)
	queryParams.Set("UserName", c.apiUser)
	queryParams.Set("Command", cmd)
	queryParams.Set("ClientIp", c.clientIP)

	for _, param := range params {
		param(queryParams)
	}

	return queryParams
}

func addParam(key, value string) func(url.Values) {
	return func(values url.Values) {
		values.Set(key, value)
	}
}
