package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://iwantmyname.com/basicauth/ddns"

// Record represents a record.
type Record struct {
	Hostname string `url:"hostname,omitempty"`
	Type     string `url:"type,omitempty"`
	Value    string `url:"value,omitempty"`
	TTL      int    `url:"ttl,omitempty"`
}

// Client iwantmyname client.
type Client struct {
	username   string
	password   string
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

// Do send a request (create/add/delete) to the API.
func (c Client) Do(ctx context.Context, record Record) error {
	values, err := querystring.Values(record)
	if err != nil {
		return err
	}

	endpoint := c.baseURL
	endpoint.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), http.NoBody)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 != 2 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status code: %d, %s", resp.StatusCode, string(data))
	}

	return nil
}
