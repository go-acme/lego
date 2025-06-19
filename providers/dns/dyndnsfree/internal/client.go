package internal

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://dynup.de/acme.php"

type Client struct {
	username string
	password string

	baseURL    string
	HTTPClient *http.Client
}

func NewClient(username, password string) (*Client, error) {
	if username == "" || password == "" {
		return nil, errors.New("credentials missing")
	}

	return &Client{
		username:   username,
		password:   password,
		baseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddTXTRecord(ctx context.Context, zone, hostname, value string) error {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}

	query := baseURL.Query()
	query.Set("username", c.username)
	query.Set("password", c.password)
	query.Set("hostname", zone)
	query.Set("add_hostname", hostname)
	query.Set("txt", value)
	baseURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), http.NoBody)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if !bytes.Equal(raw, []byte("success")) {
		return errors.New(string(raw))
	}

	return nil
}
