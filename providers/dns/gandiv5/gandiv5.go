// Package gandiv5 implements a DNS provider for solving the DNS-01 challenge using Gandi LiveDNS api.
package gandiv5

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/gandiv5/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "GANDIV5_"

	EnvAPIKey              = envNamespace + "API_KEY"
	EnvPersonalAccessToken = envNamespace + "PERSONAL_ACCESS_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const minTTL = 300

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL             string
	APIKey              string // Deprecated use PersonalAccessToken
	PersonalAccessToken string
	PropagationTimeout  time.Duration
	PollingInterval     time.Duration
	TTL                 int
	HTTPClient          *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 20*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 20*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Gandi.
// Credentials must be passed in the environment variable: GANDIV5_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	// TODO(ldez): rewrite this when APIKey will be removed.
	config := NewDefaultConfig()
	config.APIKey = env.GetOrFile(EnvAPIKey)
	config.PersonalAccessToken = env.GetOrFile(EnvPersonalAccessToken)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Gandi.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gandiv5: the configuration of the DNS provider is nil")
	}

	if config.APIKey != "" {
		log.Warn("gandiv5: API Key is deprecated, use Personal Access Token instead")
	}

	if config.APIKey == "" && config.PersonalAccessToken == "" {
		return nil, errors.New("gandiv5: credentials information are missing")
	}

	if config.TTL < minTTL {
		log.Warnf(log.LazySprintf("gandiv5: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL))

		config.TTL = minTTL
	}

	client := internal.NewClient(config.APIKey, config.PersonalAccessToken)

	if config.BaseURL != "" {
		baseURL, err := url.Parse(config.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("gandiv5: %w", err)
		}

		client.BaseURL = baseURL
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
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("gandiv5: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("gandiv5: %w", err)
	}

	_, err = d.client.AddVerificationRecord(ctx, dns01.UnFqdn(authZone), subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("gandiv5: add verification record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("gandiv5: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("gandiv5: %w", err)
	}

	record, err := d.client.GetTXTRecord(ctx, dns01.UnFqdn(authZone), subDomain)
	if err != nil {
		return fmt.Errorf("gandiv5: get TXT record (clean): %w", err)
	}

	newRecord := removeTXTRecordValue(record, info.Value, d.config.TTL)

	if len(newRecord.Values) == 0 {
		err = d.client.DeleteTXTRecord(ctx, dns01.UnFqdn(authZone), subDomain)
		if err != nil {
			return fmt.Errorf("gandiv5: delete TXT record: %w", err)
		}

		return nil
	}

	err = d.client.OverwriteTXTRecord(ctx, dns01.UnFqdn(authZone), subDomain, newRecord)
	if err != nil {
		return fmt.Errorf("gandiv5: overwrite TXT record (clean): %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func removeTXTRecordValue(record *internal.Record, value string, ttl int) *internal.Record {
	var values []string

	for _, v := range record.Values {
		if v != value || v != strconv.Quote(value) {
			continue
		}

		values = append(values, v)
	}

	return &internal.Record{TTL: ttl, Values: values}
}
