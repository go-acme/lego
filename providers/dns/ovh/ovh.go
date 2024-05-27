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
// Create a Token:          https://eu.api.ovh.com/createToken/
// Create a OAuth2 client:   https://eu.api.ovh.com/console-preview/?section=%2Fme&branch=v1#post-/me/api/oauth2/client

// Environment variables names.
const (
	envNamespace = "OVH_"

	EnvEndpoint = envNamespace + "ENDPOINT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Authenticate using application key.
const (
	EnvApplicationKey    = envNamespace + "APPLICATION_KEY"
	EnvApplicationSecret = envNamespace + "APPLICATION_SECRET"
	EnvConsumerKey       = envNamespace + "CONSUMER_KEY"
)

// Authenticate using OAuth2 client.
const (
	EnvClientID     = envNamespace + "CLIENT_ID"
	EnvClientSecret = envNamespace + "CLIENT_SECRET"
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

// OAuth2Config the OAuth2 specific configuration.
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIEndpoint string

	ApplicationKey    string
	ApplicationSecret string
	ConsumerKey       string

	OAuth2Config *OAuth2Config

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

func (c *Config) hasAppKeyAuth() bool {
	return c.ApplicationKey != "" || c.ApplicationSecret != "" || c.ConsumerKey != ""
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
	config, err := createConfigFromEnvVars()
	if err != nil {
		return nil, fmt.Errorf("ovh: %w", err)
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OVH.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ovh: the configuration of the DNS provider is nil")
	}

	if config.OAuth2Config != nil && config.hasAppKeyAuth() {
		return nil, errors.New("ovh: can't use both authentication systems (ApplicationKey and OAuth2)")
	}

	client, err := newClient(config)
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
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// Parse domain name
	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ovh: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("ovh: %w", err)
	}

	reqURL := fmt.Sprintf("/domain/zone/%s/record", authZone)
	reqData := Record{FieldType: "TXT", SubDomain: subDomain, Target: info.Value, TTL: d.config.TTL}

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
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("ovh: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ovh: could not find zone for domain %q: %w", domain, err)
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

func createConfigFromEnvVars() (*Config, error) {
	firstAppKeyEnvVar := findFirstValuedEnvVar(EnvApplicationKey, EnvApplicationSecret, EnvConsumerKey)
	firstOAuth2EnvVar := findFirstValuedEnvVar(EnvClientID, EnvClientSecret)

	if firstAppKeyEnvVar != "" && firstOAuth2EnvVar != "" {
		return nil, fmt.Errorf("can't use both %s and %s at the same time", firstAppKeyEnvVar, firstOAuth2EnvVar)
	}

	config := NewDefaultConfig()

	if firstOAuth2EnvVar != "" {
		values, err := env.Get(EnvEndpoint, EnvClientID, EnvClientSecret)
		if err != nil {
			return nil, err
		}

		config.APIEndpoint = values[EnvEndpoint]
		config.OAuth2Config = &OAuth2Config{
			ClientID:     values[EnvClientID],
			ClientSecret: values[EnvClientSecret],
		}

		return config, nil
	}

	values, err := env.Get(EnvEndpoint, EnvApplicationKey, EnvApplicationSecret, EnvConsumerKey)
	if err != nil {
		return nil, err
	}

	config.APIEndpoint = values[EnvEndpoint]

	config.ApplicationKey = values[EnvApplicationKey]
	config.ApplicationSecret = values[EnvApplicationSecret]
	config.ConsumerKey = values[EnvConsumerKey]

	return config, nil
}

func findFirstValuedEnvVar(envVars ...string) string {
	for _, envVar := range envVars {
		if env.GetOrFile(envVar) != "" {
			return envVar
		}
	}

	return ""
}

func newClient(config *Config) (*ovh.Client, error) {
	if config.OAuth2Config == nil {
		return newClientApplicationKey(config)
	}

	return newClientOAuth2(config)
}

func newClientApplicationKey(config *Config) (*ovh.Client, error) {
	if config.APIEndpoint == "" || config.ApplicationKey == "" || config.ApplicationSecret == "" || config.ConsumerKey == "" {
		return nil, errors.New("credentials are missing")
	}

	client, err := ovh.NewClient(
		config.APIEndpoint,
		config.ApplicationKey,
		config.ApplicationSecret,
		config.ConsumerKey,
	)
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	return client, nil
}

func newClientOAuth2(config *Config) (*ovh.Client, error) {
	if config.APIEndpoint == "" || config.OAuth2Config.ClientID == "" || config.OAuth2Config.ClientSecret == "" {
		return nil, errors.New("credentials are missing")
	}

	client, err := ovh.NewOAuth2Client(
		config.APIEndpoint,
		config.OAuth2Config.ClientID,
		config.OAuth2Config.ClientSecret,
	)
	if err != nil {
		return nil, fmt.Errorf("new OAuth2 client: %w", err)
	}

	return client, nil
}
