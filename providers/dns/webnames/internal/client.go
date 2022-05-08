package internal

import (
	"fmt"
	"net/http"
)

const defaultBaseURL = "https://www.webnames.ru/scripts/json_domain_zone_manager.pl"

// Client the reg.ru client.
type Client struct {
	apikey string

	BaseURL    string
	HTTPClient *http.Client
}

// NewClient Creates a reg.ru client.
func NewClient(apikey string) *Client {
	return &Client{
		apikey:     apikey,
		BaseURL:    defaultBaseURL,
		HTTPClient: http.DefaultClient,
	}
}

// RemoveTxtRecord removes a TXT record.
// https://github.com/regtime-ltd/certbot-dns-webnames/blob/master/cleanup.sh
func (c Client) RemoveTxtRecord(domain, subDomain, content string) error {
	return fmt.Errorf("can not delete")
}

// AddTXTRecord adds a TXT record.
// https://github.com/regtime-ltd/certbot-dns-webnames/blob/master/authenticator.sh
func (c Client) AddTXTRecord(domain, subDomain, content string) error {
	return fmt.Errorf("can not add")
}
