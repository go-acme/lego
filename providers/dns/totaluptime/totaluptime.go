// Package totaluptime implements a DNS provider for solving the DNS-01 challenge using Total Uptime.
package totaluptime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/totaluptime/internal"
)

// Environment variables names.
const (
	envNamespace = "TOTALUPTIME_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"

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

	zoneIDs     map[string]string
	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Total Uptime.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("totaluptime: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Total Uptime.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("totaluptime: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("totaluptime: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		zoneIDs:   make(map[string]string),
		recordIDs: make(map[string]string),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("totaluptime: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("totaluptime: %w", err)
	}

	zone, err := d.findZone(ctx, authZone)
	if err != nil {
		return fmt.Errorf("totaluptime: %w", err)
	}

	record := internal.TXTRecord{
		HostName: subDomain,
		Text:     info.Value,
		TTL:      strconv.Itoa(d.config.TTL),
	}

	result, err := d.client.AddTXTRecord(ctx, zone.ID, record)
	if err != nil {
		return fmt.Errorf("totaluptime: add TXT record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.zoneIDs[token] = zone.ID
	d.recordIDs[token] = result.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, recordOK := d.recordIDs[token]
	zoneID, zoneOK := d.zoneIDs[token]
	d.recordIDsMu.Unlock()

	if !zoneOK {
		return fmt.Errorf("totaluptime: unknown zone ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	if !recordOK {
		return fmt.Errorf("totaluptime: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	_, err := d.client.DeleteTXTRecord(ctx, zoneID, recordID)
	if err != nil {
		return fmt.Errorf("totaluptime: delete record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.zoneIDs, token)
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, authZone string) (internal.Domain, error) {
	options := internal.Options{
		SearchField:     internal.FieldDomainName,
		SearchString:    dns01.UnFqdn(authZone),
		SearchOperation: internal.SearchOperationEqual,
	}

	domains, err := d.client.GetDomains(ctx, options)
	if err != nil {
		return internal.Domain{}, fmt.Errorf("get domains: %w", err)
	}

	if len(domains) == 0 {
		return internal.Domain{}, errors.New("no domains found")
	}

	if len(domains) > 1 {
		return internal.Domain{}, errors.New("multiple domains found")
	}

	return domains[0], nil
}
