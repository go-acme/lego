// Package rainyun implements a DNS provider for solving the DNS-01 challenge using Rain Yun.
package rainyun

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/rainyun/internal"
)

// Environment variables names.
const (
	envNamespace = "RAINYUN_"

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
}

// NewDNSProvider returns a DNSProvider instance configured for Rain Yun.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("rainyun: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Rain Yun.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("rainyun: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("rainyun: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("rainyun: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("rainyun: %w", err)
	}

	domainID, err := d.findDomainID(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("rainyun: find domain ID: %w", err)
	}

	record := internal.Record{
		Host:     subDomain,
		Priority: 10,
		Line:     "DEFAULT",
		TTL:      d.config.TTL,
		Type:     "TXT",
		Value:    info.Value,
	}

	err = d.client.AddRecord(ctx, domainID, record)
	if err != nil {
		return fmt.Errorf("rainyun: add record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("rainyun: could not find zone for domain %q: %w", domain, err)
	}

	domainID, err := d.findDomainID(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("rainyun: find domain ID: %w", err)
	}

	recordID, err := d.findRecordID(ctx, domainID, info)
	if err != nil {
		return fmt.Errorf("rainyun: find record ID: %w", err)
	}

	err = d.client.DeleteRecord(ctx, domainID, recordID)
	if err != nil {
		return fmt.Errorf("rainyun: delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findDomainID(ctx context.Context, domain string) (int, error) {
	domains, err := d.client.ListDomains(ctx)
	if err != nil {
		return 0, err
	}

	for _, dom := range domains {
		if dom.Domain == domain {
			return dom.ID, nil
		}
	}

	return 0, fmt.Errorf("domain not found: %s", domain)
}

func (d *DNSProvider) findRecordID(ctx context.Context, domainID int, info dns01.ChallengeInfo) (int, error) {
	records, err := d.client.ListRecords(ctx, domainID)
	if err != nil {
		return 0, fmt.Errorf("list records: %w", err)
	}

	zone := dns01.UnFqdn(info.EffectiveFQDN)

	for _, record := range records {
		if strings.HasPrefix(zone, record.Host) && record.Value == info.Value {
			return record.ID, nil
		}
	}

	return 0, fmt.Errorf("record not found: domainID=%d, fqdn=%s", domainID, info.EffectiveFQDN)
}
