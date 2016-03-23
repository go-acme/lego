// Package dnsimple implements a DNS provider for solving the DNS-01 challenge
// using dnsimple DNS.
package dnsimple

import (
	"fmt"
	"os"
	"strings"

	"github.com/weppos/dnsimple-go/dnsimple"
	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *dnsimple.Client
}

// NewDNSProvider returns a DNSProvider instance configured for dnsimple.
// Credentials must be passed in the environment variables: DNSIMPLE_EMAIL
// and DNSIMPLE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	email := os.Getenv("DNSIMPLE_EMAIL")
	key := os.Getenv("DNSIMPLE_API_KEY")
	return NewDNSProviderCredentials(email, key)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for dnsimple.
func NewDNSProviderCredentials(email, key string) (*DNSProvider, error) {
	if email == "" || key == "" {
		return nil, fmt.Errorf("DNSimple credentials missing")
	}

	return &DNSProvider{
		client: dnsimple.NewClient(key, email),
	}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (c *DNSProvider) Present(domain *acme.Domain, token, keyAuth string) error {
	fqdn, value, ttl := domain.GetDNS01Record(keyAuth)

	zoneID, zoneName, err := c.getHostedZone(domain.Domain)
	if err != nil {
		return err
	}

	recordAttributes := c.newTxtRecord(zoneName, fqdn, value, ttl)
	_, _, err = c.client.Domains.CreateRecord(zoneID, *recordAttributes)
	if err != nil {
		return fmt.Errorf("DNSimple API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (c *DNSProvider) CleanUp(domain *acme.Domain, token, keyAuth string) error {
	fqdn, _, _ := domain.GetDNS01Record(keyAuth)

	records, err := c.findTxtRecords(domain.Domain, fqdn)
	if err != nil {
		return err
	}

	for _, rec := range records {
		_, err := c.client.Domains.DeleteRecord(rec.DomainId, rec.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *DNSProvider) getHostedZone(domain string) (string, string, error) {
	domains, _, err := c.client.Domains.List()
	if err != nil {
		return "", "", fmt.Errorf("DNSimple API call failed: %v", err)
	}

	var hostedDomain dnsimple.Domain
	for _, d := range domains {
		if strings.HasSuffix(domain, d.Name) {
			if len(d.Name) > len(hostedDomain.Name) {
				hostedDomain = d
			}
		}
	}
	if hostedDomain.Id == 0 {
		return "", "", fmt.Errorf("No matching DNSimple domain found for domain %s", domain)
	}

	return fmt.Sprintf("%v", hostedDomain.Id), hostedDomain.Name, nil
}

func (c *DNSProvider) findTxtRecords(domain, fqdn string) ([]dnsimple.Record, error) {
	zoneID, zoneName, err := c.getHostedZone(domain)
	if err != nil {
		return nil, err
	}

	var records []dnsimple.Record
	result, _, err := c.client.Domains.ListRecords(zoneID, "", "TXT")
	if err != nil {
		return records, fmt.Errorf("DNSimple API call has failed: %v", err)
	}

	recordName := c.extractRecordName(fqdn, zoneName)
	for _, record := range result {
		if record.Name == recordName {
			records = append(records, record)
		}
	}

	return records, nil
}

func (c *DNSProvider) newTxtRecord(zone, fqdn, value string, ttl int) *dnsimple.Record {
	name := c.extractRecordName(fqdn, zone)

	return &dnsimple.Record{
		Type:    "TXT",
		Name:    name,
		Content: value,
		TTL:     ttl,
	}
}

func (c *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
