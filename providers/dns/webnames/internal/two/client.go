package two

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://www.webnames.ru:81/RegTimeSRS.pl"

// Client the Webnames client.
type Client struct {
	username string
	password string

	baseURL    string
	HTTPClient *http.Client
}

// NewClient Creates a Webnames client.
func NewClient(username string, password string) *Client {
	return &Client{
		username:   username,
		password:   password,
		baseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// AddTXTRecord adds a TXT record.
func (c Client) AddTXTRecord(domain, subDomain, content string) error {
	data := url.Values{}
	data.Set("thisPage", "pispDomainZoneAddTXT")
	data.Set("domain_name", domain)
	data.Set("subname", subDomain)
	data.Set("text", content)

	return c.do(data)
}

// RemoveTxtRecord removes a TXT record.
func (c Client) RemoveTxtRecord(domain, subDomain string) error {
	data := url.Values{}
	data.Set("thisPage", "pispDomainZoneRmRR")
	data.Set("domain_name", domain)
	data.Set("subname", subDomain)
	data.Set("rectype", "TXT")

	return c.do(data)
}

func (c Client) do(data url.Values) error {
	data.Set("username", c.username)
	data.Set("password", c.password)

	req, err := http.NewRequest(http.MethodPost, c.baseURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		all, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status code: %d: %s", resp.StatusCode, string(all))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if bytes.HasPrefix(body, []byte("Error:")) {
		return errors.New(string(body))
	}

	return nil
}
