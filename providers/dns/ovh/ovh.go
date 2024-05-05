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
// Create a OAuth client:   https://eu.api.ovh.com/console-preview/?section=%2Fme&branch=v1#post-/me/api/oauth2/client

// Environment variables names.
const (
	envNamespace = "OVH_"

	EnvEndpoint = envNamespace + "ENDPOINT"

	// Authenticate using application key
	EnvApplicationKey    = envNamespace + "APPLICATION_KEY"
	EnvApplicationSecret = envNamespace + "APPLICATION_SECRET"
	EnvConsumerKey       = envNamespace + "CONSUMER_KEY"

	// Authenticate using OAuth2 client
	EnvClientId     = envNamespace + "CLIENT_ID"
	EnvClientSecret = envNamespace + "CLIENT_SECRET"

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

type ApplicationConfig struct {
	ApplicationKey    string
	ApplicationSecret string
	ConsumerKey       string
}

type OAuth2Config struct {
	ClientId     string
	ClientSecret string
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIEndpoint        string
	ApplicationConfig  *ApplicationConfig
	OAuth2Config       *OAuth2Config
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
	// If OVH_CLIENT_ID is set create an OAuth2Config variant of config
	if _, err := env.Get(EnvClientId); err == nil {
		// Both authentication are mutually exclusive
		if _, err := env.Get(EnvApplicationKey); err == nil {
			return nil, fmt.Errorf("ovh: set %v or %v but not both", EnvApplicationKey, EnvClientId)
		}

		values, err := env.Get(EnvEndpoint, EnvClientId, EnvClientSecret)
		if err != nil {
			return nil, fmt.Errorf("ovh: %w", err)
		}

		config := NewDefaultConfig()
		config.APIEndpoint = values[EnvEndpoint]
		config.OAuth2Config = &OAuth2Config{
			ClientId:     values[EnvClientId],
			ClientSecret: values[EnvClientSecret],
		}

		return NewDNSProviderOAuth2Config(config)
	}

	// Else create an ApplicationConfig variant of config
	values, err := env.Get(EnvEndpoint, EnvApplicationKey, EnvApplicationSecret, EnvConsumerKey)
	if err != nil {
		return nil, fmt.Errorf("ovh: %w", err)
	}

	config := NewDefaultConfig()
	config.APIEndpoint = values[EnvEndpoint]

	config.ApplicationConfig = &ApplicationConfig{
		ApplicationKey:    values[EnvApplicationKey],
		ApplicationSecret: values[EnvApplicationSecret],
		ConsumerKey:       values[EnvConsumerKey],
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OVH.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ovh: the configuration of the DNS provider is nil")
	}

	apiConfig := config.ApplicationConfig
	if apiConfig == nil {
		return nil, errors.New("ovh: the configuration of ApplicationConfig is nil")
	}

	if config.APIEndpoint == "" || apiConfig.ApplicationKey == "" || apiConfig.ApplicationSecret == "" || apiConfig.ConsumerKey == "" {
		return nil, errors.New("ovh: credentials missing")
	}

	client, err := ovh.NewClient(
		config.APIEndpoint,
		apiConfig.ApplicationKey,
		apiConfig.ApplicationSecret,
		apiConfig.ConsumerKey,
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

// NewDNSProviderConfig return a DNSProvider instance configured for OVH.
func NewDNSProviderOAuth2Config(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ovh: the configuration of the DNS provider is nil")
	}

	oauth2Config := config.OAuth2Config
	if oauth2Config == nil {
		return nil, errors.New("ovh: the configuration of OAuth2Config is nil")
	}

	if config.APIEndpoint == "" || oauth2Config.ClientId == "" || oauth2Config.ClientSecret == "" {
		return nil, errors.New("ovh: credentials missing")
	}

	client, err := ovh.NewOAuth2Client(
		config.APIEndpoint,
		oauth2Config.ClientId,
		oauth2Config.ClientSecret,
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
