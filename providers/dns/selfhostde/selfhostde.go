// Package selfhostde implements a DNS provider for solving the DNS-01 challenge using SelfHost.(de|eu).
package selfhostde

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
	"github.com/go-acme/lego/v4/providers/dns/selfhostde/internal"
)

// Environment variables.
const (
	envNamespace = "SELFHOSTDE_"

	EnvUsername       = envNamespace + "USERNAME"
	EnvPassword       = envNamespace + "PASSWORD"
	EnvRecordsMapping = envNamespace + "RECORDS_MAPPING"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username string
	Password string

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

func (c *Config) getSeqNext(domain string) (string, error) {
	effectiveDomain := strings.TrimPrefix(domain, "_acme-challenge.")

	c.recordsMappingMu.Lock()
	defer c.recordsMappingMu.Unlock()

	seq, ok := c.RecordsMapping[effectiveDomain]
	if !ok {
		// fallback
		seq, ok = c.RecordsMapping[domain]
		if !ok {
			return "", fmt.Errorf("record mapping not found for %q", effectiveDomain)
		}
	}

	return seq.Next(), nil
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for SelfHost.(de|eu).
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword, EnvRecordsMapping)
	if err != nil {
		return nil, fmt.Errorf("selfhostde: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	mapping, err := parseRecordsMapping(values[EnvRecordsMapping])
	if err != nil {
		return nil, fmt.Errorf("selfhostde: malformed records mapping: %w", err)
	}

	config.RecordsMapping = mapping

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for SelfHost.(de|eu).
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("selfhostde: supplied configuration is nil")
	}

	if config.Username == "" || config.Password == "" {
		return nil, errors.New("selfhostde: credentials missing")
	}

	if len(config.RecordsMapping) == 0 {
		return nil, errors.New("selfhostde: missing record mapping")
	}

	for domain, seq := range config.RecordsMapping {
		if seq == nil || len(seq.ids) == 0 {
			return nil, fmt.Errorf("selfhostde: missing record ID for %q", domain)
		}
	}

	client := internal.NewClient(config.Username, config.Password)

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

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	recordID, err := d.config.getSeqNext(dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("selfhostde: %w", err)
	}

	err = d.client.UpdateTXTRecord(context.Background(), recordID, info.Value)
	if err != nil {
		return fmt.Errorf("selfhostde: update DNS TXT record (id=%s): %w", recordID, err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record previously created.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("selfhostde: unknown record ID for %q", dns01.UnFqdn(info.EffectiveFQDN))
	}

	err := d.client.UpdateTXTRecord(context.Background(), recordID, "empty")
	if err != nil {
		return fmt.Errorf("selfhostde: emptied DNS TXT record (id=%s): %w", recordID, err)
	}

	return nil
}
