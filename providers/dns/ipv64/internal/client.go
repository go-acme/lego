package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://ipv64.net/api"

// Client the IPv64 API client.
type Client struct {
	token string

	HTTPClient *http.Client
}

type IPV64AddRecordData struct {
	Domain     string `json:"add_record"`
	Prefix     string `json:"praefix"`
	PrefixType string `json:"type"`
	Content    string `json:"content"`
}

type IPV64DelRecordData struct {
	Domain     string `json:"del_record"`
	Prefix     string `json:"praefix"`
	PrefixType string `json:"type"`
	Content    string `json:"content"`
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

func (c Client) RemoveTXTRecord(ctx context.Context, domain string) error {
	return c.UpdateTxtRecord(ctx, domain, "", true)
}

// UpdateTxtRecord Update the domains TXT record
// To update the TXT record we just need to make one simple get request.
// In IPv64 you only have one TXT record shared with the domain and all subdomains.
func (c Client) UpdateTxtRecord(ctx context.Context, domain, txt string, clear bool) error {
	endpoint, _ := url.Parse(defaultBaseURL)

	if clear {
		endpoint.Path += "/del_record"

	} else {
		endpoint.Path += "/add_record"
	}

	prefix, mainDomain, err := getPrefix(domain)

	if err != nil {
		return fmt.Errorf("the domain needs to contain at least 3 parts")
	}

	if mainDomain == "" {
		return fmt.Errorf("unable to find the main domain for: %s", domain)
	}

	var data interface{}

	if clear {
		data = IPV64DelRecordData{
			Domain:     mainDomain,
			Prefix:     prefix,
			Content:    txt,
			PrefixType: "TXT",
		}
	} else {
		data = IPV64AddRecordData{
			Domain:     mainDomain,
			Prefix:     prefix,
			Content:    txt,
			PrefixType: "TXT",
		}
	}

	marshal, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(),
		bytes.NewReader(marshal))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	// Add the token to the request header.
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

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
	if body != "\"info\":\"success\"" {
		return fmt.Errorf("request to change TXT record for IPv64 returned the following result ("+
			"%s) this does not match expectation (OK) used url [%s]", body, endpoint)
	}
	return nil
}

// IPv64 only lets you write to your subdomain.
// It must be in format subdomain.home64.de,
// not in format subsubdomain.subdomain.home64.de.
// So strip off everything that is not top 3 levels.
func getPrefix(domain string) (prefix string, mainDomain string, err error) {
	domain = dns01.UnFqdn(domain)

	splittedPartsOfDomain := strings.Split(domain, ".")
	lengthOfSplit := len(splittedPartsOfDomain)

	println(lengthOfSplit)
	if lengthOfSplit < 3 {
		return "", "", errors.New("The domain needs to contain ")
	}

	root := splittedPartsOfDomain[lengthOfSplit-1]            // de
	resultingDomain := splittedPartsOfDomain[lengthOfSplit-2] ///homeserver
	subdomain := splittedPartsOfDomain[lengthOfSplit-3]       // home64

	const SEP = "."
	completeDomain := subdomain + SEP + resultingDomain + SEP + root
	praefix := strings.Join(splittedPartsOfDomain[:lengthOfSplit-3], ".")
	return praefix, completeDomain, nil
}
