package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/miekg/dns"
)

const defaultBaseURL = "https://www.duckdns.org/update"

// Client the DuckDNS API client.
type Client struct {
	token string

	baseURL    string
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		baseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) AddTXTRecord(ctx context.Context, domain, value string) error {
	return c.UpdateTxtRecord(ctx, domain, value, false)
}

func (c *Client) RemoveTXTRecord(ctx context.Context, domain string) error {
	return c.UpdateTxtRecord(ctx, domain, "", true)
}

// UpdateTxtRecord Update the domains TXT record
// To update the TXT record we just need to make one simple get request.
// In DuckDNS you only have one TXT record shared with the domain and all subdomains.
func (c *Client) UpdateTxtRecord(ctx context.Context, domain, txt string, clearRecord bool) error {
	endpoint, _ := url.Parse(c.baseURL)

	mainDomain := getMainDomain(domain)
	if mainDomain == "" {
		return fmt.Errorf("unable to find the main domain for: %s", domain)
	}

	query := endpoint.Query()
	query.Set("domains", mainDomain)
	query.Set("token", c.token)
	query.Set("clear", strconv.FormatBool(clearRecord))
	query.Set("txt", txt)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	body := string(raw)
	if body != "OK" {
		return fmt.Errorf("request to change TXT record for DuckDNS returned the following result (%s) this does not match expectation (OK) used url [%s]", body, endpoint)
	}

	return nil
}

// DuckDNS only lets you write to your subdomain.
// It must be in format subdomain.duckdns.org,
// not in format subsubdomain.subdomain.duckdns.org.
// So strip off everything that is not top 3 levels.
func getMainDomain(domain string) string {
	domain = dns01.UnFqdn(domain)

	split := dns.Split(domain)
	if strings.HasSuffix(strings.ToLower(domain), "duckdns.org") {
		if len(split) < 3 {
			return ""
		}

		firstSubDomainIndex := split[len(split)-3]

		return domain[firstSubDomainIndex:]
	}

	return domain[split[len(split)-1]:]
}
