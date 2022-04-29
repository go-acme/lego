// Package vercel implements a DNS provider for solving the DNS-01 challenge using Vercel DNS.
package vercel

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/vercel/internal"
)

// Environment variables names.
const (
	envNamespace = "VERCEL_"

	EnvAuthToken = envNamespace + "API_TOKEN"
	EnvTeamID    = envNamespace + "TEAM_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AuthToken          string
	TeamID             string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 60*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 5*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for Vercel.
// Credentials must be passed in the environment variables: VERCEL_API_TOKEN, VERCEL_TEAM_ID.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAuthToken)
	if err != nil {
		return nil, fmt.Errorf("vercel: %w", err)
	}

	config := NewDefaultConfig()
	config.AuthToken = values[EnvAuthToken]
	config.TeamID = env.GetOrDefaultString(EnvTeamID, "")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Digital Ocean.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("vercel: the configuration of the DNS provider is nil")
	}

	if config.AuthToken == "" {
		return nil, errors.New("vercel: credentials missing")
	}

	client := internal.NewClient(config.AuthToken, config.TeamID)

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

	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("vercel: could not determine zone for domain %q: %w", domain, err)
	}

	record := internal.Record{
		Name:  fqdn,
		Type:  "TXT",
		Value: value,
		TTL:   d.config.TTL,
	}

	respData, err := d.client.CreateRecord(authZone, record)
	if err != nil {
		return fmt.Errorf("vercel: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = respData.UID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("vercel: %w", err)
	}

	// get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("vercel: unknown record ID for '%s'", fqdn)
	}

	err = d.client.DeleteRecord(authZone, recordID)
	if err != nil {
		return fmt.Errorf("vercel: %w", err)
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}
