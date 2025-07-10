package internal

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://dynamic.zoneedit.com"

// Client the ZoneEdit API client.
type Client struct {
	userID   string
	password string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(userID, password string) (*Client, error) {
	if userID == "" || password == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		userID:     userID,
		password:   password,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) CreateTXTRecord(domain, rdata string) error {
	return c.perform("txt-create.php", domain, rdata)
}

func (c *Client) DeleteTXTRecord(domain, rdata string) error {
	return c.perform("txt-delete.php", domain, rdata)
}

func (c *Client) perform(actionPath, domain, rdata string) error {
	endpoint := c.baseURL.JoinPath(actionPath)

	query := endpoint.Query()
	query.Set("host", domain)
	query.Set("rdata", rdata)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return err
	}

	return c.do(req)
}

func (c *Client) do(req *http.Request) error {
	req.SetBasicAuth(c.userID, c.password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	// TODO(ldez) only for debug must be removed.
	fmt.Println(string(raw))

	if bytes.Contains(raw, []byte("SUCCESS CODE")) {
		return nil
	}

	raw = bytes.TrimSpace(raw)

	if bytes.HasSuffix(raw, []byte(">")) {
		raw = slices.Concat(raw[:len(raw)-1], []byte("/>"))
	}

	var apiErr APIError
	err = xml.Unmarshal(raw, &apiErr)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code: %d] %w", resp.StatusCode, apiErr)
}
