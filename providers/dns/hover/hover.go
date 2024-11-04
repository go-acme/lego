// Package hover implements a DNS provider for solving the DNS-01 challenge using Hover DNS (past: "TuCows").
package hover

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/hover/internal"
)

// Environment variables names.
const (
	envNamespace = "HOVER_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username string
	Password string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
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

// NewDNSProvider returns a DNSProvider instance configured for Hover.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, err
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig returns a DNSProvider instance configured for Hover.
func NewDNSProviderConfig(config *Config) (d *DNSProvider, err error) {
	if config == nil {
		return nil, errors.New("hover: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("hover: new client: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

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

	ctx := context.Background()

	defer func() { d.client.Logout() }()

	err := d.client.Login(ctx)
	if err != nil {
		return fmt.Errorf("hover: login: %w", err)
	}

	domains, err := d.client.GetDomains(ctx)
	if err != nil {
		return fmt.Errorf("hover: get domains: %w", err)
	}

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("timewebcloud: could not find zone for domain %q: %w", domain, err)
	}

	dom, err := findDomain(domains, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("hover: find domain: %w", err)
	}

	err = d.client.AddTXTRecord(ctx, dom.ID, dns01.UnFqdn(info.EffectiveFQDN), info.Value)
	if err != nil {
		return fmt.Errorf("hover: add TXT record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	defer func() { d.client.Logout() }()

	err := d.client.Login(ctx)
	if err != nil {
		return fmt.Errorf("hover: login: %w", err)
	}

	domains, err := d.client.GetDomains(ctx)
	if err != nil {
		return fmt.Errorf("hover: get domains: %w", err)
	}

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("timewebcloud: could not find zone for domain %q: %w", domain, err)
	}

	dom, err := findDomain(domains, authZone)
	if err != nil {
		return fmt.Errorf("hover: find domain: %w", err)
	}

	records, err := d.client.GetRecords(ctx, dom.ID)
	if err != nil {
		return fmt.Errorf("hover: get records: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, dom.DomainName)
	if err != nil {
		return fmt.Errorf("hover: extract subdomain: %w", err)
	}

	record, err := findRecord(records, subDomain)
	if err != nil {
		return fmt.Errorf("hover: find record: %w", err)
	}

	err = d.client.DeleteRecord(ctx, dom.ID, record.ID)
	if err != nil {
		return fmt.Errorf("hover: delete record: %w", err)
	}

	return nil
}

func findRecord(records []internal.Record, name string) (*internal.Record, error) {
	for _, record := range records {
		if record.Name == name {
			return &record, nil
		}
	}

	return nil, fmt.Errorf("record not found: %s", name)
}

func findDomain(domains []internal.Domain, name string) (*internal.Domain, error) {
	for _, v := range domains {
		if v.DomainName == name {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("domain not found: %s", name)
}
