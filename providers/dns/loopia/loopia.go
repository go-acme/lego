// Package loopia implements a DNS provider for solving the DNS-01 challenge using loopia DNS.
package loopia

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/loopia/internal"
)

// Environment variables names.
const (
	envNamespace = "LOOPIA_"

	EnvAPIUser     = envNamespace + "API_USER"
	EnvAPIPassword = envNamespace + "API_PASSWORD"
	EnvAPIURL      = envNamespace + "API_URL"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const minTTL = 300

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type dnsClient interface {
	AddTXTRecord(ctx context.Context, domain, subdomain string, ttl int, value string) error
	RemoveTXTRecord(ctx context.Context, domain, subdomain string, recordID int) error
	GetTXTRecords(ctx context.Context, domain, subdomain string) ([]internal.RecordObj, error)
	RemoveSubdomain(ctx context.Context, domain, subdomain string) error
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	APIUser            string
	APIPassword        string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 40*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPropagationTimeout),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, time.Minute),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client dnsClient

	inProgressInfo map[string]int
	inProgressMu   sync.Mutex

	// only for testing purpose.
	findZoneByFqdn func(fqdn string) (string, error)
}

// NewDNSProvider returns a DNSProvider instance configured for Loopia.
// Credentials must be passed in the environment variables:
// LOOPIA_API_USER, LOOPIA_API_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIUser, EnvAPIPassword)
	if err != nil {
		return nil, fmt.Errorf("loopia: %w", err)
	}

	config := NewDefaultConfig()
	config.APIUser = values[EnvAPIUser]
	config.APIPassword = values[EnvAPIPassword]
	config.BaseURL = env.GetOrDefaultString(EnvAPIURL, internal.DefaultBaseURL)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Loopia.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("loopia: the configuration of the DNS provider is nil")
	}

	if config.APIUser == "" || config.APIPassword == "" {
		return nil, errors.New("loopia: credentials missing")
	}

	// Min value for TTL is 300
	if config.TTL < 300 {
		config.TTL = 300
	}

	client := internal.NewClient(config.APIUser, config.APIPassword)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	if config.BaseURL != "" {
		client.BaseURL = config.BaseURL
	}

	return &DNSProvider{
		config:         config,
		client:         client,
		findZoneByFqdn: dns01.FindZoneByFqdn,
		inProgressInfo: make(map[string]int),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	subDomain, authZone, err := d.splitDomain(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("loopia: %w", err)
	}

	ctx := context.Background()

	err = d.client.AddTXTRecord(ctx, authZone, subDomain, d.config.TTL, info.Value)
	if err != nil {
		return fmt.Errorf("loopia: failed to add TXT record: %w", err)
	}

	txtRecords, err := d.client.GetTXTRecords(ctx, authZone, subDomain)
	if err != nil {
		return fmt.Errorf("loopia: failed to get TXT records: %w", err)
	}

	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()

	for _, r := range txtRecords {
		if r.Rdata == info.Value {
			d.inProgressInfo[token] = r.RecordID
			return nil
		}
	}

	return errors.New("loopia: failed to find the stored TXT record")
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	subDomain, authZone, err := d.splitDomain(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("loopia: %w", err)
	}

	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()

	ctx := context.Background()

	err = d.client.RemoveTXTRecord(ctx, authZone, subDomain, d.inProgressInfo[token])
	if err != nil {
		return fmt.Errorf("loopia: failed to remove TXT record: %w", err)
	}

	records, err := d.client.GetTXTRecords(ctx, authZone, subDomain)
	if err != nil {
		return fmt.Errorf("loopia: failed to get TXT records: %w", err)
	}

	if len(records) > 0 {
		return nil
	}

	err = d.client.RemoveSubdomain(ctx, authZone, subDomain)
	if err != nil {
		return fmt.Errorf("loopia: failed to remove subdomain: %w", err)
	}

	return nil
}

func (d *DNSProvider) splitDomain(fqdn string) (string, string, error) {
	authZone, err := d.findZoneByFqdn(fqdn)
	if err != nil {
		return "", "", fmt.Errorf("could not find zone: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, authZone)
	if err != nil {
		return "", "", err
	}

	return subDomain, dns01.UnFqdn(authZone), nil
}
