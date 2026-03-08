// Package eurodns implements a DNS provider for solving the DNS-01 challenge using EuroDNS.
package eurodns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/eurodns/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "EURODNS_"

	EnvApplicationID = envNamespace + "APP_ID"
	EnvAPIKey        = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ApplicationID string
	APIKey        string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, internal.DefaultTTL),
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
}

// NewDNSProvider returns a DNSProvider instance configured for EuroDNS.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvApplicationID, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("eurodns: %w", err)
	}

	config := NewDefaultConfig()
	config.ApplicationID = values[EnvApplicationID]
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for EuroDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("eurodns: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.ApplicationID, config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("eurodns: %w", err)
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
		return fmt.Errorf("eurodns: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("eurodns: %w", err)
	}

	authZone = dns01.UnFqdn(authZone)

	zone, err := d.client.GetZone(ctx, authZone)
	if err != nil {
		return fmt.Errorf("eurodns: get zone: %w", err)
	}

	zone.Records = append(zone.Records, internal.Record{
		Type:  "TXT",
		Host:  subDomain,
		TTL:   internal.TTLRounder(d.config.TTL),
		RData: info.Value,
	})

	validation, err := d.client.ValidateZone(ctx, authZone, zone)
	if err != nil {
		return fmt.Errorf("eurodns: validate zone: %w", err)
	}

	if validation.Report != nil && !validation.Report.IsValid {
		return fmt.Errorf("eurodns: validation report: %w", validation.Report)
	}

	err = d.client.SaveZone(ctx, authZone, zone)
	if err != nil {
		return fmt.Errorf("eurodns: save zone: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("eurodns: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("eurodns: %w", err)
	}

	authZone = dns01.UnFqdn(authZone)

	zone, err := d.client.GetZone(ctx, authZone)
	if err != nil {
		return fmt.Errorf("eurodns: get zone: %w", err)
	}

	var recordsToKeep []internal.Record

	for _, record := range zone.Records {
		if record.Type == "TXT" && record.Host == subDomain && record.RData == info.Value {
			continue
		}

		recordsToKeep = append(recordsToKeep, record)
	}

	zone.Records = recordsToKeep

	validation, err := d.client.ValidateZone(ctx, authZone, zone)
	if err != nil {
		return fmt.Errorf("eurodns: validate zone: %w", err)
	}

	if validation.Report != nil && !validation.Report.IsValid {
		return fmt.Errorf("eurodns: validation report: %w", validation.Report)
	}

	err = d.client.SaveZone(ctx, authZone, zone)
	if err != nil {
		return fmt.Errorf("eurodns: save zone: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
