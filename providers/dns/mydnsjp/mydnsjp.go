// Package mydnsjp implements a DNS provider for solving the DNS-01
// challenge using MyDNS.jp.
package mydnsjp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// APIURL represents the API endpoint to call.
// TODO: Unexport?
const APIURL = "https://www.mydns.jp/directedit.html"

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	masterid string
	password string
	client   *http.Client
}

// NewDNSProvider returns a DNSProvider instance configured for MyDNS.jp.
// Credentials must be passed in the environment variables: MYDNSJP_MASTERID
// and MYDNSJP_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("MYDNSJP_MASTERID", "MYDNSJP_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("MyDNS.jp: %v", err)
	}

	return NewDNSProviderCredentials(values["MYDNSJP_MASTERID"], values["MYDNSJP_PASSWORD"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for MyDNS.jp.
func NewDNSProviderCredentials(masterid, password string) (*DNSProvider, error) {
	if masterid == "" || password == "" {
		return nil, errors.New("MyDNS.jp: some credentials information are missing")
	}

	return &DNSProvider{
		masterid: masterid,
		password: password,
		client:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 120 * time.Second, 2 * time.Second
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	_, value, _ := acme.DNS01Record(domain, keyAuth)
	params := url.Values{}
	params.Set("CERTBOT_DOMAIN", domain)
	params.Set("CERTBOT_VALIDATION", value)
	params.Set("EDIT_CMD", "REGIST")
	req, err := http.NewRequest(http.MethodPost, APIURL, strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(d.masterid, d.password)
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("error querying MyDNS.jp API -> %v", err)
	}
	defer resp.Body.Close()
	return err
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	_, value, _ := acme.DNS01Record(domain, keyAuth)
	params := url.Values{}
	params.Set("CERTBOT_DOMAIN", domain)
	params.Set("CERTBOT_VALIDATION", value)
	params.Set("EDIT_CMD", "DELETE")
	req, err := http.NewRequest(http.MethodPost, APIURL, strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(d.masterid, d.password)
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("error querying MyDNS.jp API -> %v", err)
	}
	defer resp.Body.Close()
	return err
}
