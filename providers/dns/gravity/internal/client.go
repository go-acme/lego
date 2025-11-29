package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	"golang.org/x/net/publicsuffix"
)

// Client the Gravity API client.
type Client struct {
	username string
	password string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(serverURL, username, password string) (*Client, error) {
	if username == "" || password == "" {
		return nil, errors.New("credentials missing")
	}

	if serverURL == "" {
		return nil, errors.New("server URL missing")
	}

	baseURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		username:   username,
		password:   password,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) Login(ctx context.Context) (*Auth, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	c.HTTPClient.Jar = jar

	login := Login{
		Username: c.username,
		Password: c.password,
	}

	endpoint := c.baseURL.JoinPath("api", "v1", "auth", "login")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, login)
	if err != nil {
		return nil, err
	}

	result := &Auth{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Me(ctx context.Context) (*UserInfo, error) {
	endpoint := c.baseURL.JoinPath("api", "v1", "auth", "me")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := &UserInfo{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, err
}

func (c *Client) GetDNSZones(ctx context.Context, name string) ([]Zone, error) {
	endpoint := c.baseURL.JoinPath("api", "v1", "dns", "zones")

	if name != "" {
		query := endpoint.Query()
		query.Set("name", name)
		endpoint.RawQuery = query.Encode()
	}

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := Zones{}

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Zones, nil
}

func (c *Client) CreateDNSRecord(ctx context.Context, zone string, record Record) error {
	endpoint := c.baseURL.JoinPath("api", "v1", "dns", "zones", "records")

	query := endpoint.Query()

	query.Set("zone", zone)
	query.Set("hostname", record.Hostname)

	// When the UID is the same as an existing one, the record is updated, else a new record is created.
	// An explicit UID is not required to create a record.
	if record.UID != "" {
		query.Set("uid", record.UID)
	}

	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) DeleteDNSRecord(ctx context.Context, zone string, record Record) error {
	endpoint := c.baseURL.JoinPath("api", "v1", "dns", "zones", "records")

	query := endpoint.Query()

	query.Set("zone", zone)
	query.Set("hostname", record.Hostname)
	query.Set("uid", record.UID)
	query.Set("type", record.Type)

	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return parseError(req, resp)
	}

	if result == nil {
		return nil
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

func newJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	var errAPI APIError

	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return &errAPI
}
