// Package stackpath implements a DNS provider for solving the DNS-01 challenge using Stackpath DNS.
// https://developer.stackpath.com/en/api/dns/
package stackpath

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	defaultBaseURL = "https://gateway.stackpath.com/dns/v1/stacks/"
	defaultAuthURL = "https://gateway.stackpath.com/identity/v1/oauth2/token"
)

// Environment variables names.
const (
	envNamespace = "STACKPATH_"

	EnvClientID     = envNamespace + "CLIENT_ID"
	EnvClientSecret = envNamespace + "CLIENT_SECRET"
	EnvStackID      = envNamespace + "STACK_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ClientID           string
	ClientSecret       string
	StackID            string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 120),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	BaseURL *url.URL
	client  *http.Client
	config  *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Stackpath.
// Credentials must be passed in the environment variables:
// STACKPATH_CLIENT_ID, STACKPATH_CLIENT_SECRET, and STACKPATH_STACK_ID.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvClientID, EnvClientSecret, EnvStackID)
	if err != nil {
		return nil, fmt.Errorf("stackpath: %w", err)
	}

	config := NewDefaultConfig()
	config.ClientID = values[EnvClientID]
	config.ClientSecret = values[EnvClientSecret]
	config.StackID = values[EnvStackID]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Stackpath.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("stackpath: the configuration of the DNS provider is nil")
	}

	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, errors.New("stackpath: credentials missing")
	}

	if config.StackID == "" {
		return nil, errors.New("stackpath: stack id missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &DNSProvider{
		BaseURL: baseURL,
		client:  getOathClient(config),
		config:  config,
	}, nil
}

func getOathClient(config *Config) *http.Client {
	oathConfig := &clientcredentials.Config{
		TokenURL:     defaultAuthURL,
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
	}

	return oathConfig.Client(context.Background())
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := d.getZones(fqdn)
	if err != nil {
		return fmt.Errorf("stackpath: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, zone.Domain)
	if err != nil {
		return fmt.Errorf("stackpath: %w", err)
	}

	record := Record{
		Name: subDomain,
		Type: "TXT",
		TTL:  d.config.TTL,
		Data: value,
	}

	return d.createZoneRecord(zone, record)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zone, err := d.getZones(fqdn)
	if err != nil {
		return fmt.Errorf("stackpath: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, zone.Domain)
	if err != nil {
		return fmt.Errorf("stackpath: %w", err)
	}

	records, err := d.getZoneRecords(subDomain, zone)
	if err != nil {
		return err
	}

	for _, record := range records {
		err = d.deleteZoneRecord(zone, record)
		if err != nil {
			log.Printf("stackpath: failed to delete TXT record: %v", err)
		}
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
