// Package ovh implements a DNS provider for solving the DNS-01 challenge using OVH DNS.
package ovh

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/ovh/go-ovh/ovh"
)

// OVH API reference:       https://eu.api.ovh.com/
// Create a Token:					https://eu.api.ovh.com/createToken/

// Environment variables names.
const (
	envNamespace = "OVH_"

	EnvEndpoint          = envNamespace + "ENDPOINT"
	EnvApplicationKey    = envNamespace + "APPLICATION_KEY"
	EnvApplicationSecret = envNamespace + "APPLICATION_SECRET"
	EnvConsumerKey       = envNamespace + "CONSUMER_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Record a DNS record.
type Record struct {
	ID        int64  `json:"id,omitempty"`
	FieldType string `json:"fieldType,omitempty"`
	SubDomain string `json:"subDomain,omitempty"`
	Target    string `json:"target,omitempty"`
	TTL       int    `json:"ttl,omitempty"`
	Zone      string `json:"zone,omitempty"`
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIEndpoint        string
	ApplicationKey     string
	ApplicationSecret  string
	ConsumerKey        string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, ovh.DefaultTimeout),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config      *Config
	client      *ovh.Client
	recordIDs   map[string]int64
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for OVH
// Credentials must be passed in the environment variables:
// OVH_ENDPOINT (must be either "ovh-eu" or "ovh-ca"), OVH_APPLICATION_KEY, OVH_APPLICATION_SECRET, OVH_CONSUMER_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvEndpoint, EnvApplicationKey, EnvApplicationSecret, EnvConsumerKey)
	if err != nil {
		return nil, fmt.Errorf("ovh: %w", err)
	}

	config := NewDefaultConfig()
	config.APIEndpoint = values[EnvEndpoint]
	config.ApplicationKey = values[EnvApplicationKey]
	config.ApplicationSecret = values[EnvApplicationSecret]
	config.ConsumerKey = values[EnvConsumerKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OVH.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ovh: the configuration of the DNS provider is nil")
	}

	if config.APIEndpoint == "" || config.ApplicationKey == "" || config.ApplicationSecret == "" || config.ConsumerKey == "" {
		return nil, errors.New("ovh: credentials missing")
	}

	client, err := ovh.NewClient(
		config.APIEndpoint,
		config.ApplicationKey,
		config.ApplicationSecret,
		config.ConsumerKey,
	)
	if err != nil {
		return nil, fmt.Errorf("ovh: %w", err)
	}

	client.Client = config.HTTPClient

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int64),
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	// Parse domain name
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("ovh: could not determine zone for domain %q: %w", fqdn, err)
	}

	authZone = dns01.UnFqdn(authZone)

	subDomain, err := dns01.ExtractSubDomain(fqdn, authZone)
	if err != nil {
		return fmt.Errorf("ovh: %w", err)
	}

	reqURL := fmt.Sprintf("/domain/zone/%s/record", authZone)
	reqData := Record{FieldType: "TXT", SubDomain: subDomain, Target: value, TTL: d.config.TTL}

	// Create TXT record
	var respData Record
	err = d.client.Post(reqURL, reqData, &respData)
	if err != nil {
		return fmt.Errorf("ovh: error when call api to add record (%s): %w", reqURL, err)
	}

	// Apply the change
	reqURL = fmt.Sprintf("/domain/zone/%s/refresh", authZone)
	err = d.client.Post(reqURL, nil, nil)
	if err != nil {
		return fmt.Errorf("ovh: error when call api to refresh zone (%s): %w", reqURL, err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = respData.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	// get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("ovh: unknown record ID for '%s'", fqdn)
	}

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("ovh: could not determine zone for domain %q: %w", fqdn, err)
	}

	authZone = dns01.UnFqdn(authZone)

	reqURL := fmt.Sprintf("/domain/zone/%s/record/%d", authZone, recordID)

	err = d.client.Delete(reqURL, nil)
	if err != nil {
		return fmt.Errorf("ovh: error when call OVH api to delete challenge record (%s): %w", reqURL, err)
	}

	// Apply the change
	reqURL = fmt.Sprintf("/domain/zone/%s/refresh", authZone)
	err = d.client.Post(reqURL, nil, nil)
	if err != nil {
		return fmt.Errorf("ovh: error when call api to refresh zone (%s): %w", reqURL, err)
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
