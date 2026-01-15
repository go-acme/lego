// Package todaynic implements a DNS provider for solving the DNS-01 challenge using TodayNIC.
package todaynic

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge/dnsnew"
	"github.com/go-acme/lego/v5/platform/config/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/todaynic/internal"
)

// Environment variables names.
const (
	envNamespace = "TODAYNIC_"

	EnvAuthUserID = envNamespace + "AUTH_USER_ID"
	EnvAPIKey     = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AuthUserID string
	APIKey     string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dnsnew.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dnsnew.DefaultPollingInterval),
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

// NewDNSProvider returns a DNSProvider instance configured for TodayNIC.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAuthUserID, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("todaynic: %w", err)
	}

	config := NewDefaultConfig()
	config.AuthUserID = values[EnvAuthUserID]
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for TodayNIC.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("todaynic: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.AuthUserID, config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("todaynic: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dnsnew.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dnsnew.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("todaynic: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dnsnew.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("todaynic: %w", err)
	}

	record := internal.Record{
		Domain: dnsnew.UnFqdn(authZone),
		Host:   subDomain,
		Type:   "TXT",
		Value:  info.Value,
		TTL:    strconv.Itoa(d.config.TTL),
	}

	recordID, err := d.client.AddRecord(ctx, record)
	if err != nil {
		return fmt.Errorf("todaynic: add record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dnsnew.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("todaynic: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecord(ctx, recordID)
	if err != nil {
		return fmt.Errorf("todaynic: delete record: %w", err)
	}

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
