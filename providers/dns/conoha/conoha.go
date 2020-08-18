// Package conoha implements a DNS provider for solving the DNS-01 challenge using ConoHa DNS.
package conoha

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/providers/dns/conoha/internal"
)

// Environment variables names.
const (
	envNamespace = "CONOHA_"

	EnvRegion      = envNamespace + "REGION"
	EnvTenantID    = envNamespace + "TENANT_ID"
	EnvAPIUsername = envNamespace + "API_USERNAME"
	EnvAPIPassword = envNamespace + "API_PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Region             string
	TenantID           string
	Username           string
	Password           string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		Region:             env.GetOrDefaultString(conf, EnvRegion, "tyo1"),
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for ConoHa DNS.
// Credentials must be passed in the environment variables:
// CONOHA_TENANT_ID, CONOHA_API_USERNAME, CONOHA_API_PASSWORD.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvTenantID, EnvAPIUsername, EnvAPIPassword)
	if err != nil {
		return nil, fmt.Errorf("conoha: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.TenantID = values[EnvTenantID]
	config.Username = values[EnvAPIUsername]
	config.Password = values[EnvAPIPassword]

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

	auth := internal.Auth{
		TenantID: config.TenantID,
		PasswordCredentials: internal.PasswordCredentials{
			Username: config.Username,
			Password: config.Password,
		},
	}

	client, err := internal.NewClient(config.Region, auth, config.HTTPClient)
	if err != nil {
		return nil, fmt.Errorf("conoha: failed to create client: %w", err)
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return err
	}

	id, err := d.client.GetDomainID(authZone)
	if err != nil {
		return fmt.Errorf("conoha: failed to get domain ID: %w", err)
	}

	record := internal.Record{
		Name: fqdn,
		Type: "TXT",
		Data: value,
		TTL:  d.config.TTL,
	}

	err = d.client.CreateRecord(id, record)
	if err != nil {
		return fmt.Errorf("conoha: failed to create record: %w", err)
	}

	return nil
}

// CleanUp clears ConoHa DNS TXT record.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return err
	}

	domID, err := d.client.GetDomainID(authZone)
	if err != nil {
		return fmt.Errorf("conoha: failed to get domain ID: %w", err)
	}

	recID, err := d.client.GetRecordID(domID, fqdn, "TXT", value)
	if err != nil {
		return fmt.Errorf("conoha: failed to get record ID: %w", err)
	}

	err = d.client.DeleteRecord(domID, recID)
	if err != nil {
		return fmt.Errorf("conoha: failed to delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
