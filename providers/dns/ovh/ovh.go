// Package ovh implements a DNS provider for solving the DNS-01 challenge using OVH DNS.
package ovh

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	"github.com/ovh/go-ovh/ovh"
)

// OVH API reference:       https://eu.api.ovh.com/
// Create a Token:          https://eu.api.ovh.com/createToken/
// Create a OAuth2 client:   https://eu.api.ovh.com/console/?section=%2Fme&branch=v1#post-/me/api/oauth2/client

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

// EnvAccessToken Authenticate using Access Token client.
const EnvAccessToken = envNamespace + "ACCESS_TOKEN"

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

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

	AccessToken string

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

func (c *Config) hasAppKeyAuth() bool {
	return c.ApplicationKey != "" || c.ApplicationSecret != "" || c.ConsumerKey != ""
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
	config := NewDefaultConfig()

	// https://github.com/ovh/go-ovh/blob/6817886d12a8c5650794b28da635af9fcdfd1162/ovh/configuration.go#L105
	config.APIEndpoint = env.GetOrDefaultString(EnvEndpoint, "ovh-eu")

	config.ApplicationKey = env.GetOrFile(EnvApplicationKey)
	config.ApplicationSecret = env.GetOrFile(EnvApplicationSecret)
	config.ConsumerKey = env.GetOrFile(EnvConsumerKey)

	config.AccessToken = env.GetOrFile(EnvAccessToken)

	clientID := env.GetOrFile(EnvClientID)
	clientSecret := env.GetOrFile(EnvClientSecret)

	if clientID != "" || clientSecret != "" {
		config.OAuth2Config = &OAuth2Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OVH.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ovh: the configuration of the DNS provider is nil")
	}

	if config.OAuth2Config != nil && config.hasAppKeyAuth() && config.AccessToken != "" {
		return nil, errors.New("ovh: can't use multiple authentication systems (ApplicationKey, OAuth2, Access Token)")
	}

	if config.OAuth2Config != nil && config.AccessToken != "" {
		return nil, errors.New("ovh: can't use multiple authentication systems (OAuth2, Access Token)")
	}

	if config.OAuth2Config != nil && config.hasAppKeyAuth() {
		return nil, errors.New("ovh: can't use multiple authentication systems (ApplicationKey, OAuth2)")
	}

	if config.hasAppKeyAuth() && config.AccessToken != "" {
		return nil, errors.New("ovh: can't use multiple authentication systems (ApplicationKey, Access Token)")
	}

	client, err := newClient(config)
	if err != nil {
		return nil, fmt.Errorf("ovh: %w", err)
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int64),
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

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

func newClient(config *Config) (*ovh.Client, error) {
	var client *ovh.Client
	var err error

	switch {
	case config.hasAppKeyAuth():
		client, err = ovh.NewClient(config.APIEndpoint, config.ApplicationKey, config.ApplicationSecret, config.ConsumerKey)
	case config.OAuth2Config != nil:
		client, err = ovh.NewOAuth2Client(config.APIEndpoint, config.OAuth2Config.ClientID, config.OAuth2Config.ClientSecret)
	case config.AccessToken != "":
		client, err = ovh.NewAccessTokenClient(config.APIEndpoint, config.AccessToken)
	default:
		client, err = ovh.NewDefaultClient()
	}

	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	client.UserAgent = useragent.Get()

	if config.HTTPClient != nil {
		client.Client = config.HTTPClient
	}

	client.Client = clientdebug.Wrap(client.Client)

	return client, nil
}
