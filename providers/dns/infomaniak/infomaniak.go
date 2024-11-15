// Package infomaniak implements a DNS provider for solving the DNS-01 challenge using Infomaniak DNS.
package infomaniak

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/infomaniak/internal"
)

// Infomaniak API reference: https://api.infomaniak.com/doc
// Create a Token: https://manager.infomaniak.com/v3/infomaniak-api

// Environment variables names.
const (
	envNamespace = "INFOMANIAK_"

	EnvEndpoint    = envNamespace + "ENDPOINT"
	EnvAccessToken = envNamespace + "ACCESS_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIEndpoint        string
	AccessToken        string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		APIEndpoint:        env.GetOrDefaultString(EnvEndpoint, internal.DefaultBaseURL),
		TTL:                env.GetOrDefaultInt(EnvTTL, 7200),
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

	domainIDs   map[string]uint64
	domainIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Infomaniak.
// Credentials must be passed in the environment variables: INFOMANIAK_ACCESS_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessToken)
	if err != nil {
		return nil, fmt.Errorf("infomaniak: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessToken = values[EnvAccessToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Infomaniak.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("infomaniak: the configuration of the DNS provider is nil")
	}

	if config.APIEndpoint == "" {
		return nil, errors.New("infomaniak: missing API endpoint")
	}

	if config.AccessToken == "" {
		return nil, errors.New("infomaniak: missing access token")
	}

	client, err := internal.New(internal.OAuthStaticAccessToken(config.HTTPClient, config.AccessToken), config.APIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("infomaniak: %w", err)
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]string),
		domainIDs: make(map[string]uint64),
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	ikDomain, err := d.client.GetDomainByName(ctx, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("infomaniak: could not get domain %q: %w", info.EffectiveFQDN, err)
	}

	d.domainIDsMu.Lock()
	d.domainIDs[token] = ikDomain.ID
	d.domainIDsMu.Unlock()

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, ikDomain.CustomerName)
	if err != nil {
		return fmt.Errorf("infomaniak: %w", err)
	}

	record := internal.Record{
		Source: subDomain,
		Target: info.Value,
		Type:   "TXT",
		TTL:    d.config.TTL,
	}

	recordID, err := d.client.CreateDNSRecord(ctx, ikDomain, record)
	if err != nil {
		return fmt.Errorf("infomaniak: error when calling api to create DNS record: %w", err)
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
		return fmt.Errorf("infomaniak: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	d.domainIDsMu.Lock()
	domainID, ok := d.domainIDs[token]
	d.domainIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("infomaniak: unknown domain ID for '%s'", info.EffectiveFQDN)
	}

	err := d.client.DeleteDNSRecord(context.Background(), domainID, recordID)
	if err != nil {
		return fmt.Errorf("infomaniak: could not delete record %q: %w", dns01.UnFqdn(info.EffectiveFQDN), err)
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	// Delete domain ID from map
	d.domainIDsMu.Lock()
	delete(d.domainIDs, token)
	d.domainIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
