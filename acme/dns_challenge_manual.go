package acme

import (
	"bufio"
	"fmt"
	"os"
)

const (
	dnsTemplate = "%s %d IN TXT \"%s\""
)

// DNSProviderManual is an implementation of the DNSProvider interface
type DNSProviderManual struct{}

// NewDNSProviderManual returns a DNSProviderManual instance.
func NewDNSProviderManual() (*DNSProviderManual, error) {
	return &DNSProviderManual{}, nil
}

// CreateTXTRecord prints instructions for manually creating the TXT record
func (*DNSProviderManual) CreateTXTRecord(fqdn, value string, ttl int) error {
	dnsRecord := fmt.Sprintf(dnsTemplate, fqdn, ttl, value)
	logf("[INFO] acme: Please create the following TXT record in your DNS zone:")
	logf("[INFO] acme: %s", dnsRecord)
	logf("[INFO] acme: Press 'Enter' when you are done")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	return nil
}

// RemoveTXTRecord prints instructions for manually removing the TXT record
func (*DNSProviderManual) RemoveTXTRecord(fqdn, value string, ttl int) error {
	dnsRecord := fmt.Sprintf(dnsTemplate, fqdn, ttl, value)
	logf("[INFO] acme: You can now remove this TXT record from your DNS zone:")
	logf("[INFO] acme: %s", dnsRecord)
	return nil
}
