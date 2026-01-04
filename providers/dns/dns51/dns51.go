// Package dns51 implements a DNS provider for solving the DNS-01 challenge using 51DNS.
package dns51

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/internal/env"
	"github.com/go-acme/lego/v5/providers/dns/dns51/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "DNS51_"

	EnvAPIKey    = envNamespace + "API_KEY"
	EnvAPISecret = envNamespace + "API_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey    string
	APISecret string

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
	domainIDs   map[string]int64
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for 51DNS.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("51dns: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.APISecret = values[EnvAPISecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for 51DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("51dns: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIKey, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("51dns: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int64),
		domainIDs: make(map[string]int64),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("51dns: could not find zone for domain %q: %w", domain, err)
	}

	zone, err := d.findZone(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("51dns: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("51dns: %w", err)
	}

	request := internal.RecordRequest{
		DomainID: zone.DomainID,
		Type:     "TXT",
		Host:     subDomain,
		Value:    info.Value,
		TTL:      d.config.TTL,
		Remark:   "Created by go-acme/lego",
	}

	record, err := d.client.CreateRecord(ctx, request)
	if err != nil {
		return fmt.Errorf("51dns: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.domainIDs[token] = zone.DomainID
	d.recordIDs[token] = record.RecordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, recordOK := d.recordIDs[token]
	domainID, domainOK := d.domainIDs[token]
	d.recordIDsMu.Unlock()

	if !recordOK {
		return fmt.Errorf("51dns: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	if !domainOK {
		return fmt.Errorf("51dns: unknown domain ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecord(ctx, domainID, recordID)
	if err != nil {
		return fmt.Errorf("51dns: delete record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.domainIDs, token)
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, authZone string) (*internal.Domain, error) {
	request := internal.DomainRequest{
		Page:     1,
		PageSize: 10,
	}

	for {
		data, err := d.client.ListDomains(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("list domains: %w", err)
		}

		for _, domain := range data.Data {
			if domain.Domain == authZone {
				return &domain, nil
			}
		}

		if len(data.Data) < request.PageSize || data.PageCount <= request.Page {
			break
		}

		request.Page++
	}

	return nil, errors.New("could not find the domain")
}
