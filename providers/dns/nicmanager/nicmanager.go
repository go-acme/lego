// Package nicmanager implements a DNS provider for solving the DNS-01 challenge using nicmanager DNS.
package nicmanager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/nicmanager/internal"
)

// Environment variables names.
const (
	envNamespace = "NICMANAGER_"

	EnvLogin    = envNamespace + "API_LOGIN"
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

const minTTL = 900

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Login     string
	Username  string
	Email     string
	Password  string
	OTPSecret string
	Mode      string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 5*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *internal.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for nicmanager.
// Credentials must be passed in the environment variables:
// NICMANAGER_API_LOGIN, NICMANAGER_API_USERNAME
// NICMANAGER_API_EMAIL
// NICMANAGER_API_PASSWORD
// NICMANAGER_API_OTP
// NICMANAGER_API_MODE.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("nicmanager: %w", err)
	}

	config := NewDefaultConfig()
	config.Password = values[EnvPassword]

	config.Mode = env.GetOrDefaultString(EnvMode, internal.ModeAnycast)
	config.Username = env.GetOrFile(EnvUsername)
	config.Login = env.GetOrFile(EnvLogin)
	config.Email = env.GetOrFile(EnvEmail)
	config.OTPSecret = env.GetOrFile(EnvOTP)

	if config.TTL < minTTL {
		return nil, fmt.Errorf("TTL must be higher than %d: %d", minTTL, config.TTL)
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for nicmanager.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("nicmanager: the configuration of the DNS provider is nil")
	}

	opts := internal.Options{
		Password: config.Password,
		OTP:      config.OTPSecret,
		Mode:     config.Mode,
	}

	switch {
	case config.Password == "":
		return nil, errors.New("nicmanager: credentials missing")
	case config.Email != "":
		opts.Email = config.Email
	case config.Login != "" && config.Username != "":
		opts.Login = config.Login
		opts.Username = config.Username
	default:
		return nil, errors.New("nicmanager: credentials missing")
	}

	client := internal.NewClient(opts)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{client: client, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	rootDomain, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("nicmanager: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	ctx := context.Background()

	zone, err := d.client.GetZone(ctx, dns01.UnFqdn(rootDomain))
	if err != nil {
		return fmt.Errorf("nicmanager: failed to get zone %q: %w", rootDomain, err)
	}

	// The way nic manager deals with record with multiple values is that they are completely different records with unique ids
	// Hence we don't check for an existing record here, but rather just create one
	record := internal.RecordCreateUpdate{
		Name:  info.EffectiveFQDN,
		Type:  "TXT",
		TTL:   d.config.TTL,
		Value: info.Value,
	}

	err = d.client.AddRecord(ctx, zone.Name, record)
	if err != nil {
		return fmt.Errorf("nicmanager: failed to create record [zone: %q, fqdn: %q]: %w", zone.Name, info.EffectiveFQDN, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	rootDomain, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("nicmanager: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	ctx := context.Background()

	zone, err := d.client.GetZone(ctx, dns01.UnFqdn(rootDomain))
	if err != nil {
		return fmt.Errorf("nicmanager: failed to get zone %q: %w", rootDomain, err)
	}

	name := dns01.UnFqdn(info.EffectiveFQDN)

	var existingRecord internal.Record
	var existingRecordFound bool
	for _, record := range zone.Records {
		if strings.EqualFold(record.Type, "TXT") && strings.EqualFold(record.Name, name) && record.Content == info.Value {
			existingRecord = record
			existingRecordFound = true
		}
	}

	if existingRecordFound {
		err = d.client.DeleteRecord(ctx, zone.Name, existingRecord.ID)
		if err != nil {
			return fmt.Errorf("nicmanager: failed to delete record [zone: %q, domain: %q]: %w", zone.Name, name, err)
		}
	}

	return fmt.Errorf("nicmanager: no record found to cleanup")
}
