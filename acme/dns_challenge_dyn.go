package acme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

var dynBaseURL = "https://api.dynect.net/REST"

type DynResponse struct {
	// One of 'success', 'failure', or 'incomplete'
	Status string `json:"status"`

	// The structure containing the actual results of the request
	Data json.RawMessage `json:"data"`

	// The ID of the job that was created in response to a request.
	JobId int `json:"job_id"`

	// A list of zero or more messages
	Messages json.RawMessage `json:"msgs"`
}

// DNSProviderDyn is an implementation of the DNSProvider interface that uses
// Dyn's Managed DNS API to manage TXT records for a domain.
type DNSProviderDyn struct {
	customerName string
	userName     string
	password     string
	token        string
}

// NewDNSProviderDyn returns a new DNSProviderDyn instance. customerName is
// the customer name of the Dyn account. userName is the user name. password is
// the password.
func NewDNSProviderDyn(customerName, userName, password string) (*DNSProviderDyn, error) {
	return &DNSProviderDyn{
		customerName: customerName,
		userName:     userName,
		password:     password,
	}, nil
}

func (d *DNSProviderDyn) sendRequest(method, resource string, payload interface{}) (*DynResponse, error) {
	url := fmt.Sprintf("%s/%s", dynBaseURL, resource)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if len(d.token) > 0 {
		req.Header.Set("Auth-Token", d.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Dyn API request failed with HTTP status code %d", resp.StatusCode)
	} else if resp.StatusCode == 307 {
		// TODO add support for HTTP 307 response and long running jobs
		return nil, fmt.Errorf("Dyn API request returned HTTP 307. This is currently unsupported")
	}

	var dynRes DynResponse
	err = json.NewDecoder(resp.Body).Decode(&dynRes)
	if err != nil {
		return nil, err
	}

	if dynRes.Status == "failure" {
		// TODO add better error handling
		return nil, fmt.Errorf("Dyn API request failed: %s", dynRes.Messages)
	}

	return &dynRes, nil
}

// Starts a new Dyn API Session. Authenticates using customerName, userName,
// password and receives a token to be used in for subsequent requests.
func (d *DNSProviderDyn) login() error {
	type creds struct {
		Customer string `json:"customer_name"`
		User     string `json:"user_name"`
		Pass     string `json:"password"`
	}

	type session struct {
		Token   string `json:"token"`
		Version string `json:"version"`
	}

	payload := &creds{Customer: d.customerName, User: d.userName, Pass: d.password}
	dynRes, err := d.sendRequest("POST", "Session", payload)
	if err != nil {
		return err
	}

	var s session
	err = json.Unmarshal(dynRes.Data, &s)
	if err != nil {
		return err
	}

	d.token = s.Token

	return nil
}

// Destroys Dyn Session
func (d *DNSProviderDyn) logout() error {
	if len(d.token) == 0 {
		// nothing to do
		return nil
	}

	url := fmt.Sprintf("%s/Session", dynBaseURL)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-Token", d.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Dyn API request failed to delete session with HTTP status code %d", resp.StatusCode)
	}

	d.token = ""

	return nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProviderDyn) Present(domain, token, keyAuth string) error {
	err := d.login()
	if err != nil {
		return err
	}

	fqdn, value, ttl := DNS01Record(domain, keyAuth)

	data := map[string]interface{}{
		"rdata": map[string]string{
			"txtdata": value,
		},
		"ttl": strconv.Itoa(ttl),
	}

	resource := fmt.Sprintf("TXTRecord/%s/%s/", domain, fqdn)
	_, err = d.sendRequest("POST", resource, data)
	if err != nil {
		return err
	}

	err = d.publish(domain, "Added TXT record for ACME dns-01 challenge using lego client")
	if err != nil {
		return err
	}

	err = d.logout()
	if err != nil {
		return err
	}

	return nil
}

func (d *DNSProviderDyn) publish(domain, notes string) error {
	type publish struct {
		Publish bool   `json:"publish"`
		Notes   string `json:"notes"`
	}

	pub := &publish{Publish: true, Notes: notes}
	resource := fmt.Sprintf("Zone/%s/", domain)
	_, err := d.sendRequest("PUT", resource, pub)
	if err != nil {
		return err
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProviderDyn) CleanUp(domain, token, keyAuth string) error {
	err := d.login()
	if err != nil {
		return err
	}

	fqdn, _, _ := DNS01Record(domain, keyAuth)

	resource := fmt.Sprintf("TXTRecord/%s/%s/", domain, fqdn)
	url := fmt.Sprintf("%s/%s", dynBaseURL, resource)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-Token", d.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Dyn API request failed to delete TXT record HTTP status code %d", resp.StatusCode)
	}

	err = d.publish(domain, "Removed TXT record for ACME dns-01 challenge using lego client")
	if err != nil {
		return err
	}

	err = d.logout()
	if err != nil {
		return err
	}

	return nil
}
