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
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/feno/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
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

// minTTL is the lowest TTL the FENO API accepts.
const minTTL = 15

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
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
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
// Credentials must be passed in the environment variable: FENO_API_KEY.
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

	if config.TTL < minTTL {
		return nil, fmt.Errorf("feno: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client, err := internal.NewClient(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("feno: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

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
		var err error

		zone, err = d.findZone(ctx, info.EffectiveFQDN)
		if err != nil {
			return fmt.Errorf("feno: could not find zone for domain %q: %w", domain, err)
		}
	}

	if !recordOK {
		var err error

		recordID, err = d.findRecordID(ctx, zone, info)
		if err != nil {
			return fmt.Errorf("feno: %w", err)
		}
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

// findZone walks the labels of the FQDN, asking the API which candidate is a zone hosted by FENO.
// The registrable domain cannot be guessed from a label count: `.no` holds both plain second-level names and delegated sub-zones.
func (d *DNSProvider) findZone(ctx context.Context, fqdn string) (string, error) {
	var lastErr error

	for candidate := range dns01.UnFqdnDomainsSeq(fqdn) {
		if !strings.Contains(candidate, ".") {
			// TLD.
			break
		}

		if strings.HasPrefix(candidate, "_") {
			// A label such as `_acme-challenge` is never a zone apex.
			continue
		}

		zone, err := d.client.GetDomain(ctx, candidate)
		if err != nil {
			var apiErr *internal.APIError
			if !errors.As(err, &apiErr) {
				return "", err
			}

			switch apiErr.StatusCode {
			case http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests:
				// A key problem looks like "domain not found" all the way up to the TLD.
				return "", err
			case http.StatusNotFound:
				lastErr = err
				continue
			default:
				return "", err
			}
		}

		if !zone.FenoNameservers {
			return "", fmt.Errorf("the zone %q does not use FENO nameservers", zone.Domain)
		}

		return zone.Domain, nil
	}

	if lastErr != nil {
		return "", fmt.Errorf("no zone found for %q: %w", fqdn, lastErr)
	}

	return "", fmt.Errorf("no zone found for %q", fqdn)
}

func (d *DNSProvider) findRecordID(ctx context.Context, zone string, info dns01.ChallengeInfo) (int, error) {
	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return 0, err
	}

	records, err := d.client.ListRecords(ctx, zone)
	if err != nil {
		return 0, fmt.Errorf("list records: %w", err)
	}

	for _, record := range records {
		if record.Type == "TXT" && strings.EqualFold(record.Name, subDomain) && record.Value == info.Value {
			return record.ID, nil
		}
	}

	return 0, fmt.Errorf("record not found for '%s' '%s'", info.EffectiveFQDN, info.Value)
}
