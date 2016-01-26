package acme

import (
	"fmt"

	"github.com/weppos/go-dnsimple/dnsimple"
)

// DNSProviderDNSimple is an implementation of the DNSProvider interface.
type DNSProviderDNSimple struct {
	client *dnsimple.Client
}

// NewDNSProviderDNSimple returns a DNSProviderDNSimple instance with a configured dnsimple client.
// Authentication is either done using the passed credentials.
func NewDNSProviderDNSimple(dnsimpleEmail, dnsimpleApiKey string) (*DNSProviderDNSimple, error) {
	if dnsimpleEmail == "" || dnsimpleApiKey == "" {
		return nil, fmt.Errorf("DNSimple credentials missing")
	}

	c := &DNSProviderDNSimple{
		client: dnsimple.NewClient(dnsimpleApiKey, dnsimpleEmail),
	}

	return c, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (c *DNSProviderDNSimple) Present(domain, token, keyAuth string) error {
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (c *DNSProviderDNSimple) CleanUp(domain, token, keyAuth string) error {
	return nil
}
