package vegadns

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	vegaClient "github.com/OpenDNS/vegadns2client"
	"github.com/xenolf/lego/acme"
)

// DNSProvider implements the acme.ChallengeProvider interface
type DNSProvider struct {
	client vegaClient.VegaDNSClient
}

// NewDNSProvider instantiates the DNSProvider implementation for vegadns
// Input: none
// Output: *DNSProvider, error
func NewDNSProvider() (*DNSProvider, error) {
	key := os.Getenv("SECRET_VEGADNS_KEY")
	secret := os.Getenv("SECRET_VEGADNS_SECRET")
	vegaDNSURL := os.Getenv("VEGADNS_URL")
	if vegaDNSURL == "" {
		return &DNSProvider{}, errors.New("VEGADNS_URL env missing")
	}
	provider := DNSProvider{}
	vega := vegaClient.NewVegaDNSClient(vegaDNSURL)
	vega.APIKey = key
	vega.APISecret = secret
	provider.client = vega
	return &provider, nil
}

// Timeout providers a timeout for the provider
// Input: none
// Output: timeout, interval time.Duration
func (r *DNSProvider) Timeout() (timeout, interval time.Duration) {
	timeout = 12 * time.Minute
	interval = 1 * time.Minute
	return
}

// Present creates a TXT record using the specified parameters
// Input: domain, token, keyAuth string
// Output: error
func (r *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	_, domainID, err := r.client.GetAuthZone(fqdn)
	if err != nil {
		return fmt.Errorf("Can't find Authorative Zone for %s in Present: %v", fqdn, err)
	}

	if err := r.client.CreateTXT(domainID, fqdn, value, 10); err != nil {
		return err
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
// Input: domain, token, keyAuth string
// Output: error
func (r *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	_, domainID, err := r.client.GetAuthZone(fqdn)
	if err != nil {
		return fmt.Errorf("Can't find Authoritative Zone for %s in CleanUp: %v", fqdn, err)
	}

	txt := strings.TrimSuffix(fqdn, ".")

	recordID, err := r.client.GetRecordID(domainID, txt, "TXT")
	if err != nil {
		return fmt.Errorf("Couldnt get Record ID in CleanUp: %s", err)
	}

	return r.client.DeleteRecord(recordID)
}
