package dns01

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
)

const (
	dnsTemplate = `%s %d IN TXT "%s"`
)

// DNSProviderManual is an implementation of the ChallengeProvider interface.
type DNSProviderManual struct{}

// NewDNSProviderManual returns a DNSProviderManual instance.
func NewDNSProviderManual() (*DNSProviderManual, error) {
	return &DNSProviderManual{}, nil
}

// Present prints instructions for manually creating the TXT record.
func (d *DNSProviderManual) Present(domain, token, keyAuth string) error {
	fqdn, value := GetRecord(domain, keyAuth)
	return d.CreateRecord(domain, token, fqdn, value)
}

// CreateRecord creates a TXT record to fulfill the DNS-01 challenge.
func (*DNSProviderManual) CreateRecord(domain, token, fqdn, value string) error {
	authZone, err := FindZoneByFqdn(fqdn)
	if err != nil {
		return err
	}

	fmt.Printf("lego: Please create the following TXT record in your %s zone:\n", authZone)
	fmt.Printf(dnsTemplate+"\n", fqdn, DefaultTTL, value)
	fmt.Printf("lego: Press 'Enter' when you are done\n")

	_, err = bufio.NewReader(os.Stdin).ReadBytes('\n')

	return err
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProviderManual) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.DeleteRecord(domain, token, fqdn, value)
}

// DeleteRecord removes a creates a TXT record from the provider.
func (d *DNSProviderManual) DeleteRecord(domain, token, fqdn, value string) error {
	authZone, err := FindZoneByFqdn(fqdn)
	if err != nil {
		return err
	}

	fmt.Printf("lego: You can now remove this TXT record from your %s zone:\n", authZone)
	fmt.Printf(dnsTemplate+"\n", fqdn, DefaultTTL, "...")

	return nil
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProviderManual) Sequential() time.Duration {
	return DefaultPropagationTimeout
}
