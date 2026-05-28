// Package infomaniak implements a DNS provider for solving the DNS-01 challenge using Infomaniak.
package infomaniak

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/infomaniak/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "INFOMANIAK_"

	EnvEndpoint    = envNamespace + "ENDPOINT" // TODO(ldez): remove in v6
	EnvAccessToken = envNamespace + "ACCESS_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIEndpoint string
	AccessToken string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		APIEndpoint:        env.GetOrDefaultString(EnvEndpoint, internal.DefaultBaseURL),
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	zones       map[string]string
	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Infomaniak.
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

	client, err := internal.NewClient(
		clientdebug.Wrap(
			internal.OAuthStaticAccessToken(config.HTTPClient, config.AccessToken),
		),
		config.APIEndpoint,
	)
	if err != nil {
		return nil, fmt.Errorf("infomaniak: new client: %w", err)
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		zones:     make(map[string]string),
		recordIDs: make(map[string]int),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zone, err := d.findZone(ctx, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("infomaniak: %w", err)
	}

	d.recordIDsMu.Lock()
	d.zones[token] = zone
	d.recordIDsMu.Unlock()

	subDomain, err := dns01.ExtractSubDomain(dns01.UnFqdn(info.EffectiveFQDN), zone)
	if err != nil {
		return fmt.Errorf("infomaniak: %w", err)
	}

	r := internal.RecordRequest{
		Source: subDomain,
		Target: info.Value,
		TTL:    d.config.TTL,
		Type:   "TXT",
	}

	record, err := d.client.CreateRecord(ctx, zone, r)
	if err != nil {
		return fmt.Errorf("infomaniak: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = record.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	zone, ok := d.zones[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("infomaniak: unknown zone for '%s' '%s'", info.EffectiveFQDN, token)
	}

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("infomaniak: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecord(ctx, zone, recordID)
	if err != nil {
		return fmt.Errorf("infomaniak: delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, fdqn string) (string, error) {
	for n := range dns01.UnFqdnDomainsSeq(fdqn) {
		exists, err := d.client.ZoneExists(ctx, n)
		if err != nil {
			return "", fmt.Errorf("check zone (%s): %w", n, err)
		}

		if exists {
			return n, nil
		}
	}

	return "", fmt.Errorf("zone not found for domain: %s", fdqn)
}
