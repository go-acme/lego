// Package transip implements a DNS provider for solving the DNS-01 challenge using TransIP.
package transip

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/transip/gotransip"
	transipdomain "github.com/transip/gotransip/domain"
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
		TTL:                int64(env.GetOrDefaultInt("TRANSIP_TTL", 10)),
		PropagationTimeout: env.GetOrDefaultSecond("TRANSIP_PROPAGATION_TIMEOUT", 10*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("TRANSIP_POLLING_INTERVAL", 10*time.Second),
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
	values, err := env.Get("TRANSIP_ACCOUNT_NAME", "TRANSIP_PRIVATE_KEY_PATH")
	if err != nil {
		return nil, fmt.Errorf("transip: %v", err)
	}

	config := NewDefaultConfig()
	config.AccountName = values["TRANSIP_ACCOUNT_NAME"]
	config.PrivateKeyPath = values["TRANSIP_PRIVATE_KEY_PATH"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials to return a DNSProvider instance configured for TransIP.
func NewDNSProviderCredentials(accountName string, privateKeyPath string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.AccountName = accountName
	config.PrivateKeyPath = privateKeyPath

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for TransIP.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("transip: the configuration of the DNS provider is nil")
	} else if config.AccountName == "" && config.PrivateKeyPath == "" {
		return nil, errors.New("transip: some credentials information are missing: TRANSIP_ACCOUNT_NAME,TRANSIP_PRIVATE_KEY_PATH")
	} else if config.AccountName == "" {
		return nil, errors.New("transip: some credentials information are missing: TRANSIP_ACCOUNT_NAME")
	} else if config.PrivateKeyPath == "" {
		return nil, errors.New("transip: some credentials information are missing: TRANSIP_PRIVATE_KEY_PATH")
	}

	client, err := gotransip.NewSOAPClient(gotransip.ClientConfig{
		AccountName:    config.AccountName,
		PrivateKeyPath: config.PrivateKeyPath,
	})
	if err != nil {
		return nil, fmt.Errorf("transip: %v", err)
	}

	return &DNSProvider{client: client, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {

	// get the FQDN
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	// get the zone
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	// get the domain name without the trailing dot
	domainName := acme.UnFqdn(authZone)

	// get the subdomain
	subdomain := strings.TrimSuffix(acme.UnFqdn(fqdn), "."+domainName)

	// get all DNS entries
	info, err := transipdomain.GetInfo(d.client, domainName)
	if err != nil {
		return fmt.Errorf("transip: error for %s in Present: %v", domain, err)
	}

	// include the new DNS entry
	dnsEntries := append(info.DNSEntries, transipdomain.DNSEntry{
		Name:    subdomain,
		TTL:     d.config.TTL,
		Type:    transipdomain.DNSEntryTypeTXT,
		Content: value,
	})

	// set the updated DNS entries
	err = transipdomain.SetDNSEntries(d.client, domainName, dnsEntries)
	if err != nil {
		return fmt.Errorf("transip: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {

	// get the FQDN
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	// get the zone
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	// get the domain name without the trailing dot
	domainName := acme.UnFqdn(authZone)

	// get the subdomain
	subdomain := strings.TrimSuffix(acme.UnFqdn(fqdn), "."+domainName)

	// get all DNS entries
	info, err := transipdomain.GetInfo(d.client, domainName)
	if err != nil {
		return fmt.Errorf("transip: error for %s in CleanUp: %v", fqdn, err)
	}

	// loop through the existing entries and remove the specific record
	updatedEntries := info.DNSEntries[:0]
	for _, e := range info.DNSEntries {
		if e.Name != subdomain {
			updatedEntries = append(updatedEntries, e)
		}
	}

	// set the updated DNS entries
	err = transipdomain.SetDNSEntries(d.client, domainName, updatedEntries)
	if err != nil {
		return fmt.Errorf("transip: couldn't get Record ID in CleanUp: %sv", err)
	}

	return nil
}
