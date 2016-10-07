// Package ultra implements a DNS provider for solving the DNS-01 challenge
// using UltraDNS Managed DNS.
package ultra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"strings"

	"net/url"

	"io/ioutil"

	"github.com/xenolf/lego/acme"
)

var ultraBaseURL = "https://restapi.ultradns.com/v2"

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// UltraDNS's Managed DNS API to manage TXT records for a domain.
type DNSProvider struct {
	username  string
	password  string
	grantType string
	token     string
}

// NewDNSProvider returns a DNSProvider instance configured for UltraDNS DNS.
// Credentials must be passed in the environment variables: ULTRA_CUSTOMER_NAME,
// ULTRA_USER_NAME and ULTRA_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	ultraUserName := os.Getenv("ULTRA_USER_NAME")
	ultraPassword := os.Getenv("ULTRA_PASSWORD")
	return NewDNSProviderCredentials(ultraUserName, ultraPassword)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for UltraDNS DNS.
func NewDNSProviderCredentials(user, pass string) (*DNSProvider, error) {
	if user == "" || pass == "" {
		return nil, fmt.Errorf("UltraDNS credentials missing, (%s)&(%s)", "ULTRA_USER_NAME", "ULTRA_PASSWORD")
	}

	return &DNSProvider{
		username:  user,
		password:  pass,
		grantType: "password",
	}, nil
}

func (d *DNSProvider) send(req *http.Request) (*http.Response, error) {
	if len(d.token) <= 0 {
		d.login()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.token))
	client := &http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
	return client.Do(req)
}

func (d *DNSProvider) getRecordStatus(resource string) (int, error) {
	endpoint := fmt.Sprintf("%s/%s", ultraBaseURL, resource)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return 0, err
	}
	resp, err := d.send(req)
	return resp.StatusCode, nil
}

func (d *DNSProvider) sendRequest(method, resource string, payload interface{}) error {
	endpoint := fmt.Sprintf("%s/%s", ultraBaseURL, resource)
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp, err := d.send(req)
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("UltraDNS API request failed with HTTP status code %d", resp.StatusCode)
	}
	if resp.StatusCode == 307 {
		// TODO add support for HTTP 307 response and long running jobs
		return fmt.Errorf("UltraDNS API returned with unsupported status code %d", resp.StatusCode)
	}
	return nil
}

// Starts a new UltraDNS API Session. Authenticates using customerName, userName,
// password and receives a token to be used in for subsequent requests.
func (d *DNSProvider) login() error {
	type session struct {
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
		ExpiresIn    string `json:"expires_in"`
	}
	endpoint := fmt.Sprintf("%s/%s", ultraBaseURL, "authorization/token")
	form := url.Values{}
	form.Add("username", d.username)
	form.Add("password", d.password)
	form.Add("grant_type", "password")
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	client := &http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	var s session
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	d.token = s.AccessToken

	return nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	type rd struct {
		TTL   int      `json:"ttl"`
		RData []string `json:"rdata"`
	}
	td := []string{}
	td = append(td, value)
	data := rd{TTL: ttl, RData: td}

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	owner := strings.Replace(acme.ToFqdn(domain), zone, "", -1)
	owner = strings.TrimRight(owner, ".")
	resource := fmt.Sprintf("/zones/%s/rrsets/TXT/%s", zone, owner)

	stat, err := d.getRecordStatus(resource)
	if err != nil {
		return err
	}

	method := "POST"
	if stat == 200 {
		method = "PUT"
	}

	return d.sendRequest(method, resource, data)
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	err := d.login()
	if err != nil {
		return err
	}
	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}
	owner := strings.Replace(domain, zone, "", -1)
	resource := fmt.Sprintf("/zones/%s/rrsets/TXT/%s", zone, owner)

	return d.sendRequest("DELETE", resource, nil)
}
