// Package omglol implements a DNS provider for solving the DNS-01 challenge using omg.lol.
package omglol

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/omglol/internal"
	"github.com/miekg/dns"
)

// Environment variables names.
const (
	envNamespace = "OMGLOL_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey string

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

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for omg.lol.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("omglol: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for omg.lol.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("omglol: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("omglol: credentials missing")
	}

	client, err := internal.NewClient(
		clientdebug.Wrap(
			internal.OAuthStaticAccessToken(config.HTTPClient, config.APIKey),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("omglol: %w", err)
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]string),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	address, name, err := splitFQDN(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("omglol: split FQDN: %w", err)
	}

	record := internal.Record{
		Name:    name,
		Content: info.Value,
		TTL:     d.config.TTL,
		Type:    "TXT",
	}

	newRecord, err := d.client.CreateRecord(ctx, address, record)
	if err != nil {
		return fmt.Errorf("omglol: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	address, _, err := splitFQDN(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("omglol: split FQDN: %w", err)
	}

	// get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("omglol: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	err = d.client.DeleteRecord(ctx, address, recordID)
	if err != nil {
		return fmt.Errorf("omglol: delete record: %w", err)
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

func splitFQDN(fdqn string) (string, string, error) {
	subDomain, err := dns01.ExtractSubDomain(fdqn, "omg.lol")
	if err != nil {
		return "", "", err
	}

	labels := dns.SplitDomainName(subDomain)

	return labels[len(labels)-1], strings.Join(labels[:len(labels)-1], "."), nil
}
