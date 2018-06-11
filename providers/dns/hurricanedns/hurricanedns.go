// Package hurricanedns implements a DNS provider for solving the DNS-01 challenge using Hurricane Electric DNS.
package hurricanedns

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/xenolf/lego/acme"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"regexp"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// dns.he.net web page to manage TXT records for a domain.
type DNSProvider struct {
	heBaseURL  string
	domainName string
	userName   string
	password   string
}

// NewDNSProvider returns a DNSProvider instance configured for Hurricane Electric DNS.
// Credentials must be passed in the environment variables: HE_DOMAIN_NAME,
// HE_USERNAME, HE_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	domainName := os.Getenv("HE_DOMAIN_NAME")
	userName := os.Getenv("HE_USERNAME")
	password := os.Getenv("HE_PASSWORD")
	return NewDNSProviderCredentials(domainName, userName, password)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for Hurricane Electric DNS.
func NewDNSProviderCredentials(domainName, userName, password string) (*DNSProvider, error) {
	if domainName == "" || userName == "" || password == "" {
		return nil, fmt.Errorf("dns.he.net credentials missing")
	}

	return &DNSProvider{
		heBaseURL:  "https://dns.he.net/",
		domainName: domainName,
		userName:   userName,
		password:   password,
	}, nil
}

// SendRequest send request, parse the HTTP response to DOM Document
func (d *DNSProvider) SendRequest(payload map[string]string) (*goquery.Document, error) {
	form := url.Values{}
	form.Add("email", d.userName)
	form.Add("pass", d.password)
	for key, value := range payload {
		form.Add(key, value)
	}

	resp, err := http.Post(d.heBaseURL, "application/x-www-form-urlencoded", bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("dns.he.net request failed with HTTP status code %d", resp.StatusCode)
	}

	body1, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body1))
	if err != nil {
		return nil, err
	}

	if doc.Find("#dns_err").Text() == "Incorrect" {
		return doc, fmt.Errorf("unable to login to dns.he.net please check username and password")
	}

	return doc, nil
}

//Resolve zone ID from HTTP Document
func (d *DNSProvider) getZoneID(zone string) (string, error) {
	payload := make(map[string]string)

	doc, err := d.SendRequest(payload)
	if err != nil {
		return "", err
	}

	zones := make(map[string]string)
	doc.Find("#domains_table tbody td:nth-child(2) img").Each(func(i int, s *goquery.Selection) {
		domain, _ := s.Attr("name")
		onclick, _ := s.Attr("onclick")
		r, _ := regexp.Compile("hosted_dns_zoneid=(\\d+)")
		id := r.FindStringSubmatch(onclick)
		zones[domain+"."] = id[1]
	})

	if val, ok := zones[zone]; ok {
		if len(val) == 0 {
			return "", fmt.Errorf("zone %s found, but zone id is not valid", zone)
		}

		return val, nil
	}

	return "", fmt.Errorf("zone %s not found", zone)
}

//Resolve record-set ID from HTTP Document
func (d *DNSProvider) getRecordSetID(zoneID string, fqdn string) (string, error) {
	payload := map[string]string{
		"hosted_dns_zoneid":   zoneID,
		"menu":                "edit_zone",
		"hosted_dns_editzone": "",
	}
	doc, err := d.SendRequest(payload)
	if err != nil {
		return "", err
	}

	records := make(map[string]string)

	doc.Find(".generictable .dns_tr").Each(func(i int, s *goquery.Selection) {
		recordId, _ := s.Attr("id")
		recordFqdn := s.Find("td:nth-child(3)").Text()
		records[recordFqdn+"."] = recordId
	})

	if val, ok := records[fqdn]; ok {
		if len(val) == 0 {
			return "", fmt.Errorf("record %s found, but record id is not valid", fqdn)
		}

		return val, nil
	}

	return "", fmt.Errorf("record %s not found", fqdn)
}

//Delete the record-set by send form post to dns.hurricanedns.net
func (d *DNSProvider) deleteRecordSet(zoneID, recordID string) error {
	payload := map[string]string{
		"menu":                  "edit_zone",
		"hosted_dns_zoneid":     zoneID,
		"hosted_dns_recordid":   recordID,
		"hosted_dns_editzone":   "1",
		"hosted_dns_delrecord":  "1",
		"hosted_dns_delconfirm": "delete",
	}

	_, err := d.SendRequest(payload)

	return err
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	if ttl < 300 {
		ttl = 300 // 300 is hurricanedns minimum value for ttl
	}

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return fmt.Errorf("unable to get zone: %s", err)
	}

	payload := map[string]string{
		"account":               "",
		"menu":                  "edit_zone",
		"Type":                  "TXT",
		"hosted_dns_zoneid":     zoneID,
		"hosted_dns_recordid":   "",
		"hosted_dns_editzone":   "1",
		"Priority":              "",
		"Name":                  fqdn,
		"Content":               value,
		"TTL":                   string(ttl),
		"hosted_dns_editrecord": "Submit",
	}

	_, err = d.SendRequest(payload)

	return err
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return err
	}

	recordID, err := d.getRecordSetID(zoneID, fqdn)
	if err != nil {
		return fmt.Errorf("unable go get record %s for zone %s: %s", fqdn, domain, err)
	}

	return d.deleteRecordSet(zoneID, recordID)
}
