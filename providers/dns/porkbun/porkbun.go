// Package porkbun implements a DNS provider for solving the DNS-01 challenge using Porkbun.
package porkbun

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/nrdcg/porkbun"
)

// Environment variables names.
const (
	envNamespace = "PORKBUN_"

	EnvSecretAPIKey = envNamespace + "SECRET_API_KEY"
	EnvAPIKey       = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const minTTL = 300

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	SecretAPIKey       string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 10*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *porkbun.Client

	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Porkbun.
// Credentials must be passed in the environment variables:
// PORKBUN_SECRET_API_KEY, PORKBUN_PAPI_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvSecretAPIKey, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("porkbun: %w", err)
	}

	config := NewDefaultConfig()
	config.SecretAPIKey = values[EnvSecretAPIKey]
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Porkbun.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("porkbun: the configuration of the DNS provider is nil")
	}

	if config.SecretAPIKey == "" || config.APIKey == "" {
		return nil, errors.New("porkbun: some credentials information are missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("porkbun: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := porkbun.New(config.SecretAPIKey, config.APIKey)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneName, hostName, err := splitDomain(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("porkbun: %w", err)
	}

	record := porkbun.Record{
		Name:    hostName,
		Type:    "TXT",
		Content: info.Value,
		TTL:     strconv.Itoa(d.config.TTL),
	}

	ctx := context.Background()

	recordID, err := d.client.CreateRecord(ctx, dns01.UnFqdn(zoneName), record)
	if err != nil {
		return fmt.Errorf("porkbun: failed to create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// gets the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("porkbun: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	zoneName, _, err := splitDomain(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("porkbun: %w", err)
	}

	ctx := context.Background()

	err = d.client.DeleteRecord(ctx, dns01.UnFqdn(zoneName), recordID)
	if err != nil {
		return fmt.Errorf("porkbun: failed to delete record: %w", err)
	}

	return nil
}

// splitDomain splits the hostname from the authoritative zone, and returns both parts.
func splitDomain(fqdn string) (string, string, error) {
	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", "", fmt.Errorf("could not find zone: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, zone)
	if err != nil {
		return "", "", err
	}

	return zone, subDomain, nil
}
