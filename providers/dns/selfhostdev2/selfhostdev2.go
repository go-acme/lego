// Package selfhostdev2 implements a DNS provider for solving the DNS-01 challenge using SelfHost.(de|eu).
package selfhostdev2

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
	"github.com/go-acme/lego/v5/providers/dns/selfhostdev2/internal"
)

// Environment variables.
const (
	envNamespace = "SELFHOSTDEV2_"

	EnvTokens         = envNamespace + "API_KEYS"
	EnvRecordsMapping = envNamespace + "RECORDS_MAPPING"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Credentials map[string]string

	RecordsMapping   map[string]*Seq
	recordsMappingMu sync.Mutex

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 4*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 30*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

func (c *Config) getSeqNext(effectiveDomain, fallback string) (int64, error) {
	c.recordsMappingMu.Lock()
	defer c.recordsMappingMu.Unlock()

	seq, ok := c.RecordsMapping[effectiveDomain]
	if !ok {
		// fallback
		seq, ok = c.RecordsMapping[fallback]
		if !ok {
			return 0, fmt.Errorf("record mapping not found for %q", effectiveDomain)
		}
	}

	return seq.Next(), nil
}

func (c *Config) getAPIKey(domain string) (string, error) {
	for s := range dns01.UnFqdnDomainsSeq(domain) {
		v, ok := c.Credentials[s]
		if !ok {
			continue
		}

		return v, nil
	}

	return "", fmt.Errorf("no API key found for %q", domain)
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	recordIDs   map[string]int64
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for SelfHost.(de|eu).
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvTokens, EnvRecordsMapping)
	if err != nil {
		return nil, fmt.Errorf("selfhostdev2: %w", err)
	}

	config := NewDefaultConfig()

	credentials, err := env.ParsePairs(values[EnvTokens])
	if err != nil {
		return nil, fmt.Errorf("selfhostdev2: credentials: %w", err)
	}

	config.Credentials = credentials

	mapping, err := parseRecordsMapping(values[EnvRecordsMapping])
	if err != nil {
		return nil, fmt.Errorf("selfhostdev2: malformed records mapping: %w", err)
	}

	config.RecordsMapping = mapping

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for SelfHost.(de|eu).
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("selfhostdev2: supplied configuration is nil")
	}

	if len(config.Credentials) == 0 {
		return nil, errors.New("selfhostdev2: credentials missing")
	}

	if len(config.RecordsMapping) == 0 {
		return nil, errors.New("selfhostdev2: missing record mapping")
	}

	for domain, seq := range config.RecordsMapping {
		if seq == nil || len(seq.ids) == 0 {
			return nil, fmt.Errorf("selfhostdev2: missing record ID for %q", domain)
		}
	}

	client := internal.NewClient()

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

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	effectiveDomain := dns01.UnFqdn(info.EffectiveDomain())
	fqdn := dns01.UnFqdn(info.EffectiveFQDN)

	recordID, err := d.config.getSeqNext(effectiveDomain, fqdn)
	if err != nil {
		return fmt.Errorf("selfhostdev2: %w", err)
	}

	apiKey, err := d.config.getAPIKey(effectiveDomain)
	if err != nil {
		return fmt.Errorf("selfhostdev2: %w", err)
	}

	payload := internal.Payload{
		APIKey:   apiKey,
		Action:   internal.ActionAdd,
		RecordID: recordID,
		Content:  info.Value,
	}

	err = d.client.UpdateTXTRecord(ctx, payload)
	if err != nil {
		return fmt.Errorf("selfhostdev2: update DNS TXT record (id=%d): %w", recordID, err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record previously created.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("selfhostdev2: unknown record ID for %q", dns01.UnFqdn(info.EffectiveFQDN))
	}

	effectiveDomain := dns01.UnFqdn(info.EffectiveDomain())

	apiKey, err := d.config.getAPIKey(effectiveDomain)
	if err != nil {
		return fmt.Errorf("selfhostdev2: %w", err)
	}

	payload := internal.Payload{
		APIKey:   apiKey,
		Action:   internal.ActionRemove,
		RecordID: recordID,
		Content:  info.Value,
	}

	err = d.client.UpdateTXTRecord(ctx, payload)
	if err != nil {
		return fmt.Errorf("selfhostdev2: emptied DNS TXT record (id=%d): %w", recordID, err)
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}
