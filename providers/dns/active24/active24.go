// Package active24 implements a DNS provider for solving the DNS-01 challenge using Active24.
package active24

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/active24"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

const baseAPIDomain = "active24.cz"

// Environment variables names.
const (
	envNamespace = "ACTIVE24_"

	EnvAPIKey = envNamespace + "API_KEY"
	EnvSecret = envNamespace + "SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey string
	Secret string

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
	client *active24.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Active24.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvSecret)
	if err != nil {
		return nil, fmt.Errorf("active24: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.Secret = values[EnvSecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Active24.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("active24: the configuration of the DNS provider is nil")
	}

	client, err := active24.NewClient(baseAPIDomain, config.APIKey, config.Secret)
	if err != nil {
		return nil, fmt.Errorf("active24: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("active24: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("active24: %w", err)
	}

	serviceID, err := d.findServiceID(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("active24: find service ID: %w", err)
	}

	record := active24.Record{
		Type:    "TXT",
		Name:    subDomain,
		Content: info.Value,
		TTL:     d.config.TTL,
	}

	err = d.client.CreateRecord(ctx, strconv.Itoa(serviceID), record)
	if err != nil {
		return fmt.Errorf("active24: create record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("active24: could not find zone for domain %q: %w", domain, err)
	}

	serviceID, err := d.findServiceID(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("active24: find service ID: %w", err)
	}

	recordID, err := d.findRecordID(ctx, strconv.Itoa(serviceID), info)
	if err != nil {
		return fmt.Errorf("active24: find record ID: %w", err)
	}

	err = d.client.DeleteRecord(ctx, strconv.Itoa(serviceID), strconv.Itoa(recordID))
	if err != nil {
		return fmt.Errorf("active24: delete record %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findServiceID(ctx context.Context, domain string) (int, error) {
	services, err := d.client.GetServices(ctx)
	if err != nil {
		return 0, fmt.Errorf("get services: %w", err)
	}

	for _, service := range services {
		if service.ServiceName != "domain" {
			continue
		}

		if service.Name != domain {
			continue
		}

		return service.ID, nil
	}

	return 0, fmt.Errorf("service not found for domain: %s", domain)
}

func (d *DNSProvider) findRecordID(ctx context.Context, serviceID string, info dns01.ChallengeInfo) (int, error) {
	// NOTE(ldez): Despite the API documentation, the filter doesn't seem to work.
	filter := active24.RecordFilter{
		Name:    dns01.UnFqdn(info.EffectiveFQDN),
		Type:    []string{"TXT"},
		Content: info.Value,
	}

	records, err := d.client.GetRecords(ctx, serviceID, filter)
	if err != nil {
		return 0, fmt.Errorf("get records: %w", err)
	}

	for _, record := range records {
		if record.Type != "TXT" {
			continue
		}

		if record.Name != dns01.UnFqdn(info.EffectiveFQDN) {
			continue
		}

		if record.Content != info.Value {
			continue
		}

		return record.ID, nil
	}

	return 0, errors.New("no record found")
}
