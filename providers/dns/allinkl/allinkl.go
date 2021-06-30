// Package allinkl implements a DNS provider for solving the DNS-01 challenge using all-inkl.
package allinkl

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/allinkl/internal"
)

// Environment variables names.
const (
	envNamespace = "ALL_INKL_"

	EnvLogin    = envNamespace + "LOGIN"
	EnvPassword = envNamespace + "PASSWORD"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Login              string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for all-inkl.
// Credentials must be passed in the environment variable: ALL_INKL_API_KEY, ALL_INKL_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvLogin, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("allinkl: %w", err)
	}

	config := NewDefaultConfig()
	config.Login = values[EnvLogin]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for all-inkl.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("allinkl: the configuration of the DNS provider is nil")
	}

	if config.Login == "" || config.Password == "" {
		return nil, errors.New("allinkl: missing credentials")
	}

	client := internal.NewClient(config.Login, config.Password)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]string),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("allinkl: could not determine zone for domain %q: %w", domain, err)
	}

	credential, err := d.client.Authentication(60, true)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}

	subDomain := dns01.UnFqdn(strings.TrimSuffix(fqdn, authZone))

	record := internal.DNSRequest{
		ZoneHost:   authZone,
		RecordType: "TXT",
		RecordName: subDomain,
		RecordData: value,
	}

	recordID, err := d.client.AddDNSSettings(credential, record)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	credential, err := d.client.Authentication(60, true)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}

	// gets the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("allinkl: unknown record ID for '%s' '%s'", fqdn, token)
	}

	_, err = d.client.DeleteDNSSettings(credential, recordID)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}

	return nil
}
