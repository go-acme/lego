// Package netlify implements a DNS provider for solving the DNS-01 challenge using Netlify.
package netlify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/netlify/internal"
)

// Environment variables names.
const (
	envNamespace = "NETLIFY_"

	EnvToken = envNamespace + "TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token              string
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

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Netlify.
// Credentials must be passed in the environment variable: NETLIFY_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken)
	if err != nil {
		return nil, fmt.Errorf("netlify: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Netlify.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("netlify: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("netlify: incomplete credentials, missing token")
	}

	client := internal.NewClient(
		clientdebug.Wrap(
			internal.OAuthStaticAccessToken(config.HTTPClient, config.Token),
		),
	)

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
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("netlify: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	record := internal.DNSRecord{
		Hostname: dns01.UnFqdn(info.EffectiveFQDN),
		TTL:      d.config.TTL,
		Type:     "TXT",
		Value:    info.Value,
	}

	resp, err := d.client.CreateRecord(context.Background(), strings.ReplaceAll(authZone, ".", "_"), record)
	if err != nil {
		return fmt.Errorf("netlify: failed to create TXT records: fqdn=%s, authZone=%s: %w", info.EffectiveFQDN, authZone, err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = resp.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("netlify: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	// gets the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("netlify: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err = d.client.RemoveRecord(context.Background(), strings.ReplaceAll(authZone, ".", "_"), recordID)
	if err != nil {
		return fmt.Errorf("netlify: failed to delete TXT records: fqdn=%s, authZone=%s, recordID=%s: %w", info.EffectiveFQDN, authZone, recordID, err)
	}

	// deletes record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}
