// Package nicmanager implements a DNS provider for solving the DNS-01 challenge using nicmanager DNS.
package nicmanager

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/nicmanager/internal"
)

// Environment variables names.
const (
	envNamespace = "NICMANAGER_"

	EnvUsername = envNamespace + "API_USERNAME"
	EnvEmail    = envNamespace + "API_EMAIL"
	EnvPassword = envNamespace + "API_PASSWORD"
	EnvOTP      = envNamespace + "API_OTP"
	EnvMode     = envNamespace + "MODE"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username  string
	Email     string
	Password  string
	OTPSecret string

	Mode string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		// Minimum allowed TTL is 900
		TTL: env.GetOrDefaultInt(EnvTTL, 900),
		// Propagation takes around 4 minutes from my testing with anycast
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 5*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *internal.NicManagerClient
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for nicmanager.
// Credentials must be passed in the environment variables: nicmanager_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("nicmanager: %w", err)
	}

	config := NewDefaultConfig()
	config.Password = values[EnvPassword]

	config.Mode = env.GetOrDefaultString(EnvMode, "anycast")
	config.Username = env.GetOrDefaultString(EnvUsername, "")
	config.Email = env.GetOrDefaultString(EnvEmail, "")
	config.OTPSecret = env.GetOrDefaultString(EnvOTP, "")

	if config.TTL < 900 {
		return nil, errors.New("minimum allowed TTL is 900")
	}
	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for nicmanager.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("nicmanager: the configuration of the DNS provider is nil")
	}

	if config.Username == "" && config.Email == "" {
		return nil, errors.New("nicmanager: credentials missing")
	}
	client := internal.NewNicManagerClient(config.HTTPClient)
	if config.Username != "" {
		if !strings.Contains(config.Username, ".") {
			return nil, fmt.Errorf("nicmanager: username '%s' must be formatted like account.user", config.Username)
		}
		parts := strings.SplitN(config.Username, ".", 1)
		client.SetAccount(parts[0], parts[1])
	} else if config.Email != "" {
		client.SetEmail(config.Email)
	}
	if config.OTPSecret != "" {
		client.SetOTP(config.OTPSecret)
	}
	client.Password = config.Password
	client.Mode = config.Mode
	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	rootDoamin, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return err
	}

	zone, err := d.client.ZoneInfo(dns01.UnFqdn(rootDoamin))
	if err != nil {
		return fmt.Errorf("nicmanager: %w", err)
	}

	// The way nic manager deals with record with multiple values is that
	// they are completely different records with unique ids
	// Hence we don't check for an existing record here, but rather just create one
	log.Infof("Create a new record for [zone: %s, fqdn: %s, domain: %s]", zone.Name, fqdn, domain)

	record := internal.RecordCreateUpdate{
		Name:  fqdn,
		Type:  "TXT",
		TTL:   d.config.TTL,
		Value: value,
	}

	err = d.client.ResourceRecordCreate(zone.Name, record)
	if err != nil {
		return fmt.Errorf("nicmanager: failed to create record [zone: %q, fqdn: %q]: %w", zone.Name, fqdn, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	rootDoamin, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return err
	}
	zone, err := d.client.ZoneInfo(dns01.UnFqdn(rootDoamin))
	if err != nil {
		return fmt.Errorf("nicmanager: %w", err)
	}

	name := dns01.UnFqdn(fqdn)

	var existingRecord internal.Record
	var existingRecordFound bool
	for _, record := range zone.Records {
		if strings.EqualFold(record.Type, "txt") && strings.EqualFold(record.Name, name) && record.Content == value {
			existingRecord = record
			existingRecordFound = true
		}
	}

	if existingRecordFound {
		err = d.client.ResourceRecordDelete(zone.Name, existingRecord.ID)
		if err != nil {
			return fmt.Errorf("nicmanager: failed to delete record [zone: %q, domain: %q]: %w", zone.Name, name, err)
		}
	}
	return fmt.Errorf("nicmanager: no record found to cleanup")
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
