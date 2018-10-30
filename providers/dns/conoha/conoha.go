// Package conoha implements a DNS provider for solving the DNS-01 challenge
// using ConoHa DNS.
package conoha

import (
	"errors"
	"fmt"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Region             string
	TenantID           string
	Username           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		Region:             env.GetOrDefaultString("CONOHA_REGION", "tyo1"),
		PropagationTimeout: env.GetOrDefaultSecond("CONOHA_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("CONOHA_POLLING_INTERVAL", acme.DefaultPollingInterval),
		TTL:                env.GetOrDefaultInt("CONOHA_TTL", 60),
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config *Config
	client *Client
}

// NewDNSProvider returns a DNSProvider instance configured for ConoHa DNS.
// Credentials must be passed in the environment variables: CONOHA_TENANT_ID, CONOHA_API_USERNAME, CONOHA_API_PASSWORD
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("CONOHA_TENANT_ID", "CONOHA_API_USERNAME", "CONOHA_API_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("conoha: %v", err)
	}

	config := NewDefaultConfig()
	config.TenantID = values["CONOHA_TENANT_ID"]
	config.Username = values["CONOHA_API_USERNAME"]
	config.Password = values["CONOHA_API_PASSWORD"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ConoHa DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("conoha: the configuration of the DNS provider is nil")
	}
	if config.TenantID == "" || config.Username == "" || config.Password == "" {
		return nil, errors.New("conoha: some credentials information are missing")
	}

	client, err := NewClient(config.Region, config.TenantID, config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("conoha: failed to login: %v", err)
	}

	return &DNSProvider{config, client}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	id, err := d.client.GetDomainID(acme.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("conoha: failed to get domain ID: %v", err)
	}

	err = d.client.CreateRecord(id, fqdn, "TXT", value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("conoha: failed to create record: %v", err)
	}

	return nil
}

// CleanUp clears ConoHa DNS TXT record
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	domID, err := d.client.GetDomainID(acme.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("conoha: failed to get domain ID: %v", err)
	}

	recID, err := d.client.GetRecordID(domID, fqdn, "TXT", value)
	if err != nil {
		return fmt.Errorf("conoha: failed to get record ID: %v", err)
	}

	err = d.client.DeleteRecord(domID, recID)
	if err != nil {
		return fmt.Errorf("conoha: failed to delete record: %v", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
