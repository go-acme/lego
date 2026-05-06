// Package xinnet implements a DNS provider for solving the DNS-01 challenge using Xinnet.
package xinnet

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
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/xinnet/internal"
)

// Environment variables names.
const (
	envNamespace = "XINNET_"

	EnvSecret  = envNamespace + "SECRET"
	EnvAgentID = envNamespace + "AGENT_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Secret  string
	AgentID string

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
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	recordIDs   map[string]int64
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Xinnet.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvSecret, EnvAgentID)
	if err != nil {
		return nil, fmt.Errorf("xinnet: %w", err)
	}

	config := NewDefaultConfig()
	config.Secret = values[EnvSecret]
	config.AgentID = values[EnvAgentID]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Xinnet.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("xinnet: the configuration of the DNS provider is nil")
	}

	signer, err := internal.NewSigner(config.Secret, config.AgentID)
	if err != nil {
		return nil, fmt.Errorf("xinnet: %w", err)
	}

	client, err := internal.NewClient(signer)
	if err != nil {
		return nil, fmt.Errorf("xinnet: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int64),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("xinnet: could not find zone for domain %q: %w", domain, err)
	}

	record := internal.Record{
		DomainName: dns01.UnFqdn(authZone),
		RecordName: dns01.UnFqdn(info.EffectiveFQDN),
		Type:       "TXT",
		Value:      info.Value,
		Line:       "默认",
		TTL:        d.config.TTL,
	}

	recordID, err := d.client.CreateRecord(ctx, record)
	if err != nil {
		return fmt.Errorf("xinnet: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("xinnet: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("xinnet: could not find zone for domain %q: %w", domain, err)
	}

	err = d.client.DeleteRecord(ctx, dns01.UnFqdn(authZone), recordID)
	if err != nil {
		return fmt.Errorf("xinnet: delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
