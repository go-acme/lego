// Package plesk implements a DNS provider for solving the DNS-01 challenge using Plesk DNS.
package plesk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/plesk/internal"
)

// Environment variables names.
const (
	envNamespace = "PLESK_"

	EnvServerBaseURL = envNamespace + "SERVER_BASE_URL"
	EnvUsername      = envNamespace + "USERNAME"
	EnvPassword      = envNamespace + "PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	baseURL  string
	Username string
	Password string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
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

	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Plesk.
// Credentials must be passed in the environment variables:
// PLESK_USERNAME and PLESK_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvServerBaseURL, EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("plesk: %w", err)
	}

	config := NewDefaultConfig()
	config.baseURL = values[EnvServerBaseURL]
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Plesk.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("plesk: the configuration of the DNS provider is nil")
	}

	if config.baseURL == "" {
		return nil, errors.New("plesk: missing server base URL")
	}

	baseURL, err := url.Parse(config.baseURL)
	if err != nil {
		return nil, fmt.Errorf("plesk: failed to parse base URL (%s): %w", config.baseURL, err)
	}

	if config.Username == "" || config.Password == "" {
		return nil, errors.New("plesk: incomplete credentials, missing username and/or password")
	}

	client := internal.NewClient(baseURL, config.Username, config.Password)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: map[string]int{},
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("plesk: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	ctx := context.Background()

	siteID, err := d.client.GetSite(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("plesk: failed to get site: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("nodion: %w", err)
	}

	recordID, err := d.client.AddRecord(ctx, siteID, subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("plesk: failed to add record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("plesk: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	_, err := d.client.DeleteRecord(context.Background(), recordID)
	if err != nil {
		return fmt.Errorf("plesk: failed to delete record (%d): %w", recordID, err)
	}

	return nil
}
