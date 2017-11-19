package nicru

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
)

var baseURL = "https://api.nic.ru"

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// DNSMadeEasy's DNS API to manage TXT records for a domain.
type DNSProvider struct {
	baseURL     string
	apiUsername string
	apiPassword string
	apiID       string
	apiSecret   string
	serviceName string
	auth        *authResponse
	client      http.Client
}

type authResponse struct {
	ExpiresIn   string `json:"expires_in"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type authError struct {
	Error string `json:"error"`
}

// Record holds the nic.ru API representation of a DNS Record
type Record struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	SourceID int    `json:"sourceId"`
}

// NewDNSProvider returns a DNSProvider instance configured for nic.ru DNS
func NewDNSProvider() (*DNSProvider, error) {
	nicruUsername := os.Getenv("NICRU_USERNAME")
	nicruPassword := os.Getenv("NICRU_PASSWORD")
	nicruAPIKey := os.Getenv("NICRU_API_KEY")
	nicruAPISecret := os.Getenv("NICRU_API_SECRET")
	nicruServiceName := os.Getenv("NICRU_SERVICE_NAME")
	provider := &DNSProvider{
		baseURL: baseURL,

		apiUsername: nicruUsername,
		apiPassword: nicruPassword,
		apiID:       nicruAPIKey,
		apiSecret:   nicruAPISecret,
		serviceName: nicruServiceName,

		client: http.Client{Timeout: 30 * time.Second},
	}

	err := provider.getAuthRespone()
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// getCredentials uses the supplied credentials to return a
// DNSProvider instance configured for nic.ru
func (d *DNSProvider) getAuthRespone() error {
	// TODO: validate all credential are set
	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("username", d.apiUsername)
	form.Add("password", d.apiPassword)
	form.Add("scope", "(GET|PUT|POST|DELETE):/dns-master/.+")
	form.Add("client_id", d.apiID)
	form.Add("client_secret", d.apiSecret)

	reqURL := fmt.Sprintf("%s/oauth/token", d.baseURL)
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("nicru auth failed, code=%v with unreadable body (%s)", resp.StatusCode, err)
		}

		var errInfo authError
		err = json.Unmarshal([]byte(body), &errInfo)
		if err != nil {
			return fmt.Errorf("nicru auth failed, code=%v without parsable error: %s, body: %s", resp.StatusCode, err, body)
		}
		return fmt.Errorf("nicru auth failed, code=%d, message=%s", resp.StatusCode, errInfo.Error)
	}

	var auth authResponse
	err = json.NewDecoder(resp.Body).Decode(&auth)
	if err != nil {
		return err
	}
	d.auth = &auth
	return nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	if ttl > 60 {
		ttl = 60
	}

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("Could not determine zone for domain: '%s'. %s", domain, err)
	}

	authZone = acme.UnFqdn(authZone)

	// http PUT https://api.nic.ru/dns-master/services/${NICRU_SERVICE_NAME}/zones/${ZONE_NAME}/records Authorization:"Bearer $TOKEN" @txt-record.xml

	name := extractRecordName(fqdn, authZone)

	reqTxt := &nicRequest{Records: []nicRecord{
		{
			Name:      name,
			TTL:       ttl,
			Type:      "TXT",
			TxtString: value,
		},
	}}
	reqTxtBody, err := xml.Marshal(reqTxt)
	if err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/dns-master/services/%s/zones/%s/records",
		d.baseURL, d.serviceName, authZone)

	req, err := http.NewRequest("PUT", reqURL, bytes.NewReader(reqTxtBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.auth.AccessToken))
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("TXT creation failed, code=%d with unreadable body (%s)", resp.StatusCode, err)
		}

		return fmt.Errorf("TXT creation failed, code=%d, body: %s", resp.StatusCode, body)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("TXT creation success, code=%d with unreadable body. error: %s", resp.StatusCode, err)
	}

	// Everything looks good, but try to decode response (should not fail in success case)
	// MAYBE: keep record from ID later to delete the record
	var respData nicResponse
	err = xml.Unmarshal(body, &respData)
	if err != nil {
		return err
	}

	return d.commitChanges(authZone)
}

func extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}

// CleanUp removes the TXT records matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("Could not determine zone for domain: '%s'. %s", domain, err)
	}
	authZone = acme.UnFqdn(authZone)

	// find records
	reqURL := fmt.Sprintf("%s/dns-master/services/%s/zones/%s/records",
		d.baseURL, d.serviceName, authZone)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.auth.AccessToken))

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("get records failed, code=%d with unreadable body (%s)", resp.StatusCode, err)
		}

		return fmt.Errorf("get records failed, code=%d, body: %s", resp.StatusCode, body)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("get records success, code=%d with unreadable body. error: %s", resp.StatusCode, err)
	}

	var respData nicResponse
	err = xml.Unmarshal(body, &respData)
	if err != nil {
		return err
	}

	name := extractRecordName(fqdn, authZone)
	var records []nicRecordResponse
	for _, rr := range respData.Zone.Records {
		if rr.Type != "TXT" {
			continue
		}
		if rr.Name == name {
			records = append(records, rr)
		}
	}

	// remove records
	for _, rr := range records {
		err = d.delRecord(authZone, rr.ID)
		if err != nil {
			return err
		}
	}

	return d.commitChanges(authZone)
}

func (d *DNSProvider) delRecord(authZone, id string) error {
	delURI := fmt.Sprintf("%s/dns-master/services/%s/zones/%s/records/%s",
		d.baseURL, d.serviceName, authZone, id)
	req, err := http.NewRequest("DELETE", delURI, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.auth.AccessToken))

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("del record with id=%s failed, code=%d with unreadable body (%s)",
				id, resp.StatusCode, err)
		}
		return fmt.Errorf("del record with id=%s, code=%d, body: %s", id, resp.StatusCode, body)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 120 * time.Second, 10 * time.Second
}

// http -v POST https://api.nic.ru/dns-master/services/${NICRU_SERVICE_NAME}/zones/${ZONE_NAME}/commit Authorization:"Bearer $TOKEN"
func (d *DNSProvider) commitChanges(authZone string) error {
	delURI := fmt.Sprintf("%s/dns-master/services/%s/zones/%s/commit",
		d.baseURL, d.serviceName, authZone)
	req, err := http.NewRequest("POST", delURI, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.auth.AccessToken))

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("commit failed, code=%d with unreadable body (%s)",
				resp.StatusCode, err)
		}
		return fmt.Errorf("commit failed, code=%d, body: %s", resp.StatusCode, body)
	}
	return nil
}
