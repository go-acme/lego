// Package feno implements a DNS provider for solving the DNS-01 challenge using FENO.
package feno

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
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/feno/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/hashicorp/go-retryablehttp"
)

// Environment variables names.
const (
	envNamespace = "FENO_"

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

	zones       map[string]string
	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for FENO.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("feno: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for FENO.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("feno: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("feno: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.HTTPClient = client.HTTPClient
	retryClient.Logger = log.Default()

	client.HTTPClient = clientdebug.Wrap(retryClient.StandardClient())

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

	zone, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("feno: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("feno: %w", err)
	}

	record := internal.Record{
		Type:  "TXT",
		Name:  subDomain,
		Value: info.Value,
		TTL:   d.config.TTL,
	}

	newRecord, err := d.client.CreateRecord(ctx, zone, record)
	if err != nil {
		return fmt.Errorf("feno: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.zones[token] = zone
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	zone, zoneOK := d.zones[token]
	recordID, recordOK := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !zoneOK {
		return fmt.Errorf("feno: unknown zone for '%s' '%s'", info.EffectiveFQDN, token)
	}

	if !recordOK {
		return fmt.Errorf("feno: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecord(ctx, zone, recordID)
	if err != nil {
		return fmt.Errorf("feno: delete record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.zones, token)
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, fqdn string) (string, error) {
	var lastErr error

	for domain := range dns01.UnFqdnDomainsSeq(fqdn) {
		// TLD.
		if !strings.Contains(domain, ".") {
			break
		}

		zone, err := d.client.GetDomain(ctx, domain)
		if err != nil {
			var apiErr *internal.APIError
			if !errors.As(err, &apiErr) {
				return "", err
			}

			if apiErr.StatusCode == http.StatusNotFound {
				lastErr = err
				continue
			}

			return "", err
		}

		if !zone.FenoNS {
			return "", fmt.Errorf("the zone %q does not use FENO nameservers", zone.Domain)
		}

		return zone.Domain, nil
	}

	if lastErr != nil {
		return "", fmt.Errorf("no zone found for %q: %w", fqdn, lastErr)
	}

	return "", fmt.Errorf("no zone found for %q", fqdn)
}
