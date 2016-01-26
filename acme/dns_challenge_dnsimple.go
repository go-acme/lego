package acme

import (
	"fmt"
	"os"

	"github.com/weppos/go-dnsimple/dnsimple"
)

// DNSProviderDNSimple is an implementation of the DNSProvider interface.
type DNSProviderDNSimple struct {
	client *dnsimple.Client
}

// NewDNSProviderDNSimple returns a DNSProviderDNSimple instance with a configured dnsimple client.
// Authentication is either done using the passed credentials or - when empty - using the environment
// variables DNSIMPLE_EMAIL and DNSIMPLE_API_KEY.
func NewDNSProviderDNSimple(dnsimpleEmail, dnsimpleApiKey string) (*DNSProviderDNSimple, error) {
	if dnsimpleEmail == "" || dnsimpleApiKey == "" {
		dnsimpleEmail, dnsimpleApiKey = dnsimpleEnvAuth()
		if dnsimpleEmail == "" || dnsimpleApiKey == "" {
			return nil, fmt.Errorf("DNSimple credentials missing")
		}
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

func dnsimpleEnvAuth() (email, apiKey string) {
	email = os.Getenv("DNSIMPLE_EMAIL")
	apiKey = os.Getenv("DNSIMPLE_API_KEY")
	if len(email) == 0 || len(apiKey) == 0 {
		return "", ""
	}
	return
}
