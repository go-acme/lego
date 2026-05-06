package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://dinahosting.com/special/api.php"

// Client the Dinahosting API client.
type Client struct {
	username string
	password string

	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(username, password string) (*Client, error) {
	if username == "" || password == "" {
		return nil, errors.New("credentials missing")
	}

	return &Client{
		username:   username,
		password:   password,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddTXTRecord(ctx context.Context, record TXTRecord) error {
	req, err := c.newRequest(ctx, "Domain_Zone_AddTypeTXT", record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) DeleteTXTRecord(ctx context.Context, record TXTRecord) error {
	req, err := c.newRequest(ctx, "Domain_Zone_DeleteTypeTXT", record)
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

func (c *Client) newRequest(ctx context.Context, command string, data any) (*http.Request, error) {
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	values, err := querystring.Values(data)
	if err != nil {
		return nil, fmt.Errorf("unable to encode parameters: %w", err)
	}

	values.Set("AUTH_USER", c.username)
	values.Set("AUTH_PWD", c.password)
	values.Set("responseType", "Json")
	values.Set("command", command)

	endpoint.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
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
