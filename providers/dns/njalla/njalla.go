// Package njalla implements a DNS provider for solving the DNS-01 challenge using Njalla.
package njalla

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
	"github.com/go-acme/lego/v4/providers/dns/njalla/internal"
	"github.com/miekg/dns"
)

// Environment variables names.
const (
	envNamespace = "NJALLA_"

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

// NewDNSProvider returns a DNSProvider instance configured for Njalla.
// Credentials must be passed in the environment variable: NJALLA_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken)
	if err != nil {
		return nil, fmt.Errorf("njalla: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Njalla.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("njalla: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("njalla: missing credentials")
	}

	client := internal.NewClient(config.Token)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

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

	rootDomain, subDomain, err := splitDomain(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("njalla: %w", err)
	}

	record := internal.Record{
		Name:    subDomain,                // TODO need to be tested
		Domain:  dns01.UnFqdn(rootDomain), // TODO need to be tested
		Content: info.Value,
		TTL:     d.config.TTL,
		Type:    "TXT",
	}

	resp, err := d.client.AddRecord(context.Background(), record)
	if err != nil {
		return fmt.Errorf("njalla: failed to add record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = resp.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	rootDomain, _, err := splitDomain(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("njalla: %w", err)
	}

	// gets the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("njalla: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err = d.client.RemoveRecord(context.Background(), recordID, dns01.UnFqdn(rootDomain))
	if err != nil {
		return fmt.Errorf("njalla: failed to delete TXT records: fqdn=%s, recordID=%s: %w", info.EffectiveFQDN, recordID, err)
	}

	// deletes record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

func splitDomain(full string) (string, string, error) {
	split := dns.Split(full)
	if len(split) < 2 {
		return "", "", fmt.Errorf("unsupported domain: %s", full)
	}

	if len(split) == 2 {
		return full, "", nil
	}

	domain := full[split[len(split)-2]:]
	subDomain := full[:split[len(split)-2]-1]

	return domain, subDomain, nil
}
