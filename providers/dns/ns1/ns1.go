// Package ns1 implements a DNS provider for solving the DNS-01 challenge using NS1 DNS.
package ns1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dnsnew"
	"github.com/go-acme/lego/v5/platform/config/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

// Environment variables names.
const (
	envNamespace = "NS1_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dnsnew.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dnsnew.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dnsnew.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *rest.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for NS1.
// Credentials must be passed in the environment variables: NS1_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("ns1: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for NS1.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ns1: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("ns1: credentials missing")
	}

	if config.HTTPClient == nil {
		// Because the rest.NewClient uses the http.DefaultClient.
		config.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}

	client := rest.NewClient(clientdebug.Wrap(config.HTTPClient), rest.SetAPIKey(config.APIKey))

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dnsnew.GetChallengeInfo(ctx, domain, keyAuth)

	zone, err := d.getHostedZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ns1: %w", err)
	}

	record, _, err := d.client.Records.Get(zone.Zone, dnsnew.UnFqdn(info.EffectiveFQDN), "TXT")

	// Create a new record
	if errors.Is(err, rest.ErrRecordMissing) || record == nil {
		// Work through a bug in the NS1 API library that causes 400 Input validation failed (Value None for field '<obj>.filters' is not of type ...)
		// So the `tags` and `blockedTags` parameters should be initialized to empty.
		record = dns.NewRecord(zone.Zone, dnsnew.UnFqdn(info.EffectiveFQDN), "TXT", make(map[string]string), make([]string, 0))
		record.TTL = d.config.TTL
		record.Answers = []*dns.Answer{{Rdata: []string{info.Value}}}

		_, err = d.client.Records.Create(record)
		if err != nil {
			return fmt.Errorf("ns1: create record [zone: %q, fqdn: %q]: %w", zone.Zone, info.EffectiveFQDN, err)
		}

		return nil
	}

	if err != nil {
		return fmt.Errorf("ns1: get the existing record: %w", err)
	}

	// Update the existing records
	record.Answers = append(record.Answers, &dns.Answer{Rdata: []string{info.Value}})

	_, err = d.client.Records.Update(record)
	if err != nil {
		return fmt.Errorf("ns1: update record [zone: %q, fqdn: %q]: %w", zone.Zone, info.EffectiveFQDN, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dnsnew.GetChallengeInfo(ctx, domain, keyAuth)

	zone, err := d.getHostedZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ns1: %w", err)
	}

	name := dnsnew.UnFqdn(info.EffectiveFQDN)

	_, err = d.client.Records.Delete(zone.Zone, name, "TXT")
	if err != nil {
		return fmt.Errorf("ns1: delete record [zone: %q, domain: %q]: %w", zone.Zone, name, err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getHostedZone(ctx context.Context, fqdn string) (*dns.Zone, error) {
	authZone, err := dnsnew.DefaultClient().FindZoneByFqdn(ctx, fqdn)
	if err != nil {
		return nil, fmt.Errorf("could not find zone: %w", err)
	}

	authZone = dnsnew.UnFqdn(authZone)

	zone, _, err := d.client.Zones.Get(authZone, false)
	if err != nil {
		return nil, fmt.Errorf("get zone [authZone: %q, fqdn: %q]: %w", authZone, fqdn, err)
	}

	return zone, nil
}
