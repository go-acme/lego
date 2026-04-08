// Package combell implements a DNS provider for solving the DNS-01 challenge using Combell DNS.
package combell

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/combell/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

const (
	minTTL = 60
	maxTTL = 8640
)

// Environment variables names.
const (
	envNamespace = "COMBELL_"

	EnvAPIKey    = envNamespace + "API_KEY"
	EnvAPISecret = envNamespace + "API_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	APISecret          string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 3600),
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

// NewDNSProvider returns a DNSProvider instance configured for Combell DNS.
// Credentials must be passed in the environment variables:
// COMBELL_API_KEY, COMBELL_API_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("combell: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.APISecret = values[EnvAPISecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Combell DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("combell: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" || config.APISecret == "" {
		return nil, errors.New("combell: some credentials information are missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("combell: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	if config.TTL > maxTTL {
		return nil, fmt.Errorf("combell: invalid TTL, TTL (%d) must be lower than %d", config.TTL, maxTTL)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}

	if config.HTTPClient != nil {
		httpClient = config.HTTPClient
	}

	client := internal.NewClient(config.APIKey, config.APISecret, clientdebug.Wrap(httpClient))

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("combell: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	record := internal.Record{
		Type:       "TXT",
		RecordName: dns01.UnFqdn(strings.TrimSuffix(info.EffectiveFQDN, authZone)),
		Content:    info.Value,
		TTL:        d.config.TTL,
	}

	err = d.client.CreateRecord(context.Background(), authZone, record)
	if err != nil {
		return fmt.Errorf("combell: create record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("combell: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	request := &internal.GetRecordsRequest{
		Type:       "TXT",
		RecordName: dns01.UnFqdn(strings.TrimSuffix(info.EffectiveFQDN, authZone)),
	}

	records, err := d.client.GetRecords(ctx, authZone, request)
	if err != nil {
		return fmt.Errorf("combell: get records: %w", err)
	}

	for _, record := range records {
		if record.Content == info.Value {
			err = d.client.DeleteRecord(ctx, authZone, record.ID)
			if err != nil {
				return fmt.Errorf("combell: delete record: %w", err)
			}
		}
	}

	return nil
}
