package internal

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://iwantmyname.com/basicauth/ddns"

// Client iwantmyname client.
type Client struct {
	username string
	password string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(username string, password string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		username:   username,
		password:   password,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// SendRequest send a request (create/add/delete) to the API.
func (c Client) SendRequest(ctx context.Context, record Record) error {
	values, err := querystring.Values(record)
	if err != nil {
		return err
	}

	endpoint := c.baseURL
	endpoint.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	return nil
}
