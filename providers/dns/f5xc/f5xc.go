// Package f5xc implements a DNS provider for solving the DNS-01 challenge using F5 XC.
package f5xc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/platform/wait"
	"github.com/go-acme/lego/v4/providers/dns/f5xc/internal"
)

// Environment variables names.
const (
	envNamespace = "F5XC_"

	EnvToken      = envNamespace + "API_TOKEN"
	EnvTenantName = envNamespace + "TENANT_NAME"
	EnvGroupName  = envNamespace + "GROUP_NAME"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIToken   string
	TenantName string
	GroupName  string

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
}

// NewDNSProvider returns a DNSProvider instance configured for F5 XC.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken, EnvTenantName, EnvGroupName)
	if err != nil {
		return nil, fmt.Errorf("f5xc: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvToken]
	config.TenantName = values[EnvTenantName]
	config.GroupName = values[EnvGroupName]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for F5 XC.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("f5xc: the configuration of the DNS provider is nil")
	}

	if config.GroupName == "" {
		return nil, errors.New("f5xc: missing group name")
	}

	client, err := internal.NewClient(config.APIToken, config.TenantName)
	if err != nil {
		return nil, fmt.Errorf("f5xc: %w", err)
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
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("f5xc: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("f5xc: %w", err)
	}

	existingRRSet, err := d.client.GetRRSet(ctx, dns01.UnFqdn(authZone), d.config.GroupName, subDomain, "TXT")
	if err != nil {
		return fmt.Errorf("f5xc: get RR Set: %w", err)
	}

	// New RRSet.
	if existingRRSet == nil || existingRRSet.RRSet.TXTRecord == nil {
		rrSet := internal.RRSet{
			Description: "lego",
			TTL:         d.config.TTL,
			TXTRecord: &internal.TXTRecord{
				Name:   subDomain,
				Values: []string{info.Value},
			},
		}

		return d.waitFor(ctx, func() error {
			_, err = d.client.CreateRRSet(ctx, dns01.UnFqdn(authZone), d.config.GroupName, rrSet)
			if err != nil {
				return fmt.Errorf("create RR set: %w", err)
			}

			return nil
		})
	}

	// Update RRSet.
	existingRRSet.RRSet.TXTRecord.Values = append(existingRRSet.RRSet.TXTRecord.Values, info.Value)

	return d.waitFor(ctx, func() error {
		_, err = d.client.ReplaceRRSet(ctx, dns01.UnFqdn(authZone), d.config.GroupName, subDomain, "TXT", existingRRSet.RRSet)
		if err != nil {
			return fmt.Errorf("replace RR set: %w", err)
		}

		return nil
	})
}

func (d *DNSProvider) waitFor(ctx context.Context, operation func() error) error {
	err := wait.Retry(ctx, operation,
		backoff.WithBackOff(backoff.NewConstantBackOff(2*time.Second)),
		backoff.WithMaxElapsedTime(60*time.Second),
	)
	if err != nil {
		return fmt.Errorf("f5xc: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("f5xc: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("f5xc: %w", err)
	}

	_, err = d.client.DeleteRRSet(context.Background(), dns01.UnFqdn(authZone), d.config.GroupName, subDomain, "TXT")
	if err != nil {
		return fmt.Errorf("f5xc: delete RR set: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
