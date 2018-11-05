// Package transip implements a DNS provider for solving the DNS-01
// challenge using TransIP.
package transip

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/transip/gotransip"
	transip_domain "github.com/transip/gotransip/domain"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	AccountName        string
	PrivateKeyPath     string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int64
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                int64(env.GetOrDefaultInt("TRANSIP_TTL", 300)),
		PropagationTimeout: env.GetOrDefaultSecond("TRANSIP_PROPAGATION_TIMEOUT", 12*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("TRANSIP_POLLING_INTERVAL", 1*time.Minute),
	}
}

// DNSProvider describes a provider for TransIP
type DNSProvider struct {
	config *Config
	client gotransip.SOAPClient
}

// NewDNSProvider returns a DNSProvider instance configured for TransIP.
// Credentials must be passed in the environment variables:
// TRANSIP_ACCOUNTNAME, TRANSIP_PRIVATEKEYPATH.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.AccountName = env.GetOrFile("TRANSIP_ACCOUNT_NAME")
	config.PrivateKeyPath = env.GetOrFile("TRANSIP_PRIVATE_KEY_PATH")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for TransIP.
// Deprecated
func NewDNSProviderCredentials(accountname string, privatekeypath string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.AccountName = accountname
	config.PrivateKeyPath = privatekeypath

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for TransIP.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("transip: the configuration of the DNS provider is nil")
	}

	c, err := gotransip.NewSOAPClient(gotransip.ClientConfig{
		AccountName: config.AccountName,
		PrivateKeyPath: config.PrivateKeyPath,
	})
	if err != nil {
		return nil, fmt.Errorf("transip: %v", err.Error())
	}

	return &DNSProvider{client: c, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	// get all DNS entries
	domainName, err := transip_domain.GetInfo(d.client, domain)

	if err != nil {
		return fmt.Errorf("transip: error for %s in Present: %v", domain, err)
	}

	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	// get the domain with the main domain
	name := strings.TrimSuffix(fqdn, "." + domain + ".")

	// append the new DNS entry
	dnsEntries := append(domainName.DNSEntries, transip_domain.DNSEntry{name, d.config.TTL, transip_domain.DNSEntryTypeTXT, value})

	err = transip_domain.SetDNSEntries(d.client, domain, dnsEntries)
	if err != nil {
		return fmt.Errorf("transip: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	// get the non-fqdn name
	name := strings.TrimSuffix(fqdn, "." + domain + ".")

	// get all DNS entries
	domainName, err := transip_domain.GetInfo(d.client, fqdn)
	if err != nil {
		return fmt.Errorf("transip: error for %s in CleanUp: %v", fqdn, err)
	}

	// create a slice with the same underlying array
	newEntries := domainName.DNSEntries[:0]

	// loop through the existing entries and remove the specific record
	for _, e := range domainName.DNSEntries {
		if  e.Name != name {
			newEntries = append(newEntries, e)
		}
	}

	err = transip_domain.SetDNSEntries(d.client, fqdn, newEntries)
	if err != nil {
		return fmt.Errorf("transip: couldn't get Record ID in CleanUp: %s", err)
	}

	return nil
}
