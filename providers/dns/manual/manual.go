package manual

import (
	"github.com/go-acme/lego/v4/challenge/dns01"
)

// DNSProvider is an implementation of the ChallengeProvider interface.
type DNSProvider = dns01.DNSProviderManual

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	return &DNSProvider{}, nil
}
