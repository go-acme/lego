package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/miekg/dns"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://ipv64.net/api"

// Client the IPv64 API client.
type Client struct {
	token string

	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c Client) AddTXTRecord(ctx context.Context, domain, value string) error {
	return c.UpdateTxtRecord(ctx, domain, value, false)
}

func (c Client) RemoveTXTRecord(ctx context.Context, domain string, value string) error {
	return c.UpdateTxtRecord(ctx, domain, value, true)
}

type SuccessMessage struct {
	Info      string `json:"info"`
	Status    string `json:"status"`
	AddRecord string `json:"add_record"`
}

// UpdateTxtRecord Update the domains TXT record
// To update the TXT record we just need to make one simple get request.
// In IPv64 you only have one TXT record shared with the domain and all subdomains.
func (c Client) UpdateTxtRecord(ctx context.Context, domain, txt string, clear bool) error {
	endpoint, _ := url.Parse(defaultBaseURL)

	prefix, mainDomain, err := getPrefix(domain)

	if err != nil {
		return fmt.Errorf("the domain needs to contain at least 3 parts")
	}

	if mainDomain == "" {
		return fmt.Errorf("unable to find the main domain for: %s", domain)
	}

	form := url.Values{}
	form.Add("praefix", prefix)
	form.Add("type", "TXT")
	form.Add("content", txt)

	if err != nil {
		return err
	}

	var req *http.Request
	var requestError error

	if clear {
		form.Add("del_record", mainDomain)
		req, requestError = http.NewRequestWithContext(ctx, http.MethodDelete, endpoint.String(), strings.NewReader(form.Encode()))
	} else {
		form.Add("add_record", mainDomain)
		req, requestError = http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(form.Encode()))
	}

	if requestError != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	// Add the token to the request header.
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)

	var successBody SuccessMessage

	body := string(raw)

	println("Content", txt)
	if parse_error := json.Unmarshal(raw, &successBody); parse_error != nil {
		return fmt.Errorf("request to change TXT record for IPv64 returned the following result ("+
			"%s) this does not match expectation (OK) used url [%s]", body, endpoint)
	}

	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if !strings.Contains(successBody.Status, "201 Created") && resp.StatusCode > 300 {
		return fmt.Errorf("request to change TXT record for IPv64 returned the following result ("+
			"%s) this does not match expectation (OK) used url [%s]", body, endpoint)
	}
	return nil
}

// IPv64 only lets you write to your subdomain.
// It must be in format subdomain.home64.de,
// not in format subsubdomain.subdomain.home64.de.
// So strip off everything that is not top 3 levels.
func getPrefix(full string) (prefix string, mainDomain string, err error) {
	split := dns.Split(full)
	if len(split) < 3 {
		return "", "", fmt.Errorf("unsupported domain: %s", full)
	}
	if len(split) == 3 {
		return "", full, nil
	}
	domain := full[split[len(split)-3]:]
	subDomain := full[:split[len(split)-3]-1]
	return subDomain, domain, nil
}
