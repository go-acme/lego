// Package bluecat implements a DNS provider for solving the DNS-01 challenge using a self-hosted Bluecat Address Manager.
package bluecat

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/bluecat/internal"
)

// Environment variables names.
const (
	envNamespace = "BLUECAT_"

	EnvServerURL  = envNamespace + "SERVER_URL"
	EnvUserName   = envNamespace + "USER_NAME"
	EnvPassword   = envNamespace + "PASSWORD"
	EnvConfigName = envNamespace + "CONFIG_NAME"
	EnvDNSView    = envNamespace + "DNS_VIEW"
	EnvDebug      = envNamespace + "DEBUG"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	UserName           string
	Password           string
	ConfigName         string
	DNSView            string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
	Debug              bool
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
		Debug: env.GetOrDefaultBool(EnvDebug, false),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Bluecat DNS.
// Credentials must be passed in the environment variables:
//	- BLUECAT_SERVER_URL
//	  It should have the scheme, hostname, and port (if required) of the authoritative Bluecat BAM server.
//	  The REST endpoint will be appended.
//	- BLUECAT_USER_NAME and BLUECAT_PASSWORD
//	- BLUECAT_CONFIG_NAME (the Configuration name)
//	- BLUECAT_DNS_VIEW (external DNS View Name)
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvServerURL, EnvUserName, EnvPassword, EnvConfigName, EnvDNSView)
	if err != nil {
		return nil, fmt.Errorf("bluecat: %w", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values[EnvServerURL]
	config.UserName = values[EnvUserName]
	config.Password = values[EnvPassword]
	config.ConfigName = values[EnvConfigName]
	config.DNSView = values[EnvDNSView]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Bluecat DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("bluecat: the configuration of the DNS provider is nil")
	}

	if config.BaseURL == "" || config.UserName == "" || config.Password == "" || config.ConfigName == "" || config.DNSView == "" {
		return nil, errors.New("bluecat: credentials missing")
	}

	client := internal.NewClient(config.BaseURL)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record using the specified parameters
// This will *not* create a sub-zone to contain the TXT record,
// so make sure the FQDN specified is within an existent zone.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	err := d.client.Login(d.config.UserName, d.config.Password)
	if err != nil {
		return fmt.Errorf("bluecat: login: %w", err)
	}

	viewID, err := d.client.LookupViewID(d.config.ConfigName, d.config.DNSView)
	if err != nil {
		return fmt.Errorf("bluecat: lookupViewID: %w", err)
	}

	parentZoneID, name, err := d.client.LookupParentZoneID(viewID, fqdn)
	if err != nil {
		return fmt.Errorf("bluecat: lookupParentZoneID: %w", err)
	}

	if d.config.Debug {
		log.Infof("fqdn: %s; viewID: %d; ZoneID: %d; zone: %s", fqdn, viewID, parentZoneID, name)
	}

	txtRecord := internal.Entity{
		Name:       name,
		Type:       internal.TXTType,
		Properties: fmt.Sprintf("ttl=%d|absoluteName=%s|txt=%s|", d.config.TTL, fqdn, value),
	}

	_, err = d.client.AddEntity(parentZoneID, txtRecord)
	if err != nil {
		return fmt.Errorf("bluecat: add TXT record: %w", err)
	}

	err = d.client.Deploy(parentZoneID)
	if err != nil {
		return fmt.Errorf("bluecat: deploy: %w", err)
	}

	err = d.client.Logout()
	if err != nil {
		return fmt.Errorf("bluecat: logout: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	err := d.client.Login(d.config.UserName, d.config.Password)
	if err != nil {
		return fmt.Errorf("bluecat: login: %w", err)
	}

	viewID, err := d.client.LookupViewID(d.config.ConfigName, d.config.DNSView)
	if err != nil {
		return fmt.Errorf("bluecat: lookupViewID: %w", err)
	}

	parentZoneID, name, err := d.client.LookupParentZoneID(viewID, fqdn)
	if err != nil {
		return fmt.Errorf("bluecat: lookupParentZoneID: %w", err)
	}

	txtRecord, err := d.client.GetEntityByName(parentZoneID, name, internal.TXTType)
	if err != nil {
		return fmt.Errorf("bluecat: get TXT record: %w", err)
	}

	err = d.client.Delete(txtRecord.ID)
	if err != nil {
		return fmt.Errorf("bluecat: delete TXT record: %w", err)
	}

	err = d.client.Deploy(parentZoneID)
	if err != nil {
		return fmt.Errorf("bluecat: deploy: %w", err)
	}

	err = d.client.Logout()
	if err != nil {
		return fmt.Errorf("bluecat: logout: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
