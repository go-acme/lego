// Package ultradns implements a DNS provider for solving the DNS-01 challenge using ultradns.
package ultradns

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/ultradns/ultradns-go-sdk/pkg/client"
	"github.com/ultradns/ultradns-go-sdk/pkg/record"
	"github.com/ultradns/ultradns-go-sdk/pkg/rrset"
)

// Environment variables names.
const (
	envNamespace = "ULTRADNS_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"
	EnvEndpoint = envNamespace + "ENDPOINT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"

	// Default variables names.
	defaultEndpoint  = "https://api.ultradns.com/"
	defaultUserAgent = "lego-provider-ultradns"
)

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *client.Client
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username string
	Password string
	Endpoint string

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		Endpoint:           env.GetOrDefaultString(EnvEndpoint, defaultEndpoint),
		TTL:                env.GetOrDefaultInt(EnvTTL, 120),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 4*time.Second),
	}
}

// NewDNSProvider returns a DNSProvider instance configured for ultradns.
// Credentials must be passed in the environment variables:
// ULTRADNS_USERNAME and ULTRADNS_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("ultradns: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ultradns.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ultradns: the configuration of the DNS provider is nil")
	}

	ultraConfig := client.Config{
		Username:  config.Username,
		Password:  config.Password,
		HostURL:   config.Endpoint,
		UserAgent: defaultUserAgent,
	}

	uClient, err := client.NewClient(ultraConfig)
	if err != nil {
		return nil, fmt.Errorf("ultradns: %w", err)
	}

	return &DNSProvider{config: config, client: uClient}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("ultradns: %w", err)
	}

	recordService, err := record.Get(d.client)
	if err != nil {
		return fmt.Errorf("ultradns: %w", err)
	}

	rrSetKeyData := &rrset.RRSetKey{
		Owner:      fqdn,
		Zone:       authZone,
		RecordType: "TXT",
	}

	res, _, _ := recordService.Read(rrSetKeyData)

	rrSetData := &rrset.RRSet{
		OwnerName: fqdn,
		TTL:       d.config.TTL,
		RRType:    "TXT",
		RData:     []string{value},
	}

	if res != nil && res.StatusCode == 200 {
		_, err = recordService.Update(rrSetKeyData, rrSetData)
	} else {
		_, err = recordService.Create(rrSetKeyData, rrSetData)
	}
	if err != nil {
		return fmt.Errorf("ultradns: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("ultradns: %w", err)
	}

	recordService, err := record.Get(d.client)
	if err != nil {
		return fmt.Errorf("ultradns: %w", err)
	}

	rrSetKeyData := &rrset.RRSetKey{
		Owner:      fqdn,
		Zone:       authZone,
		RecordType: "TXT",
	}

	_, err = recordService.Delete(rrSetKeyData)
	if err != nil {
		return fmt.Errorf("ultradns: %w", err)
	}

	return nil
}
