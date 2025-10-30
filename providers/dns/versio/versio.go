// Package versio implements a DNS provider for solving the DNS-01 challenge using versio DNS.
package versio

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/versio/internal"
)

// Environment variables names.
const (
	envNamespace = "VERSIO_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"
	EnvEndpoint = envNamespace + "ENDPOINT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            *url.URL
	TTL                int
	Username           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	baseURL, err := url.Parse(env.GetOrDefaultString(EnvEndpoint, internal.DefaultBaseURL))
	if err != nil {
		baseURL, _ = url.Parse(internal.DefaultBaseURL)
	}

	return &Config{
		BaseURL:            baseURL,
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 5*time.Second),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, dns01.DefaultPropagationTimeout),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	dnsEntriesMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("versio: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Versio.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("versio: the configuration of the DNS provider is nil")
	}

	if config.Username == "" {
		return nil, errors.New("versio: the versio username is missing")
	}

	if config.Password == "" {
		return nil, errors.New("versio: the versio password is missing")
	}

	client := internal.NewClient(config.Username, config.Password)

	if config.BaseURL != nil {
		client.BaseURL = config.BaseURL
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

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
		return fmt.Errorf("versio: could not find zone for domain %q: %w", domain, err)
	}

	// use mutex to prevent race condition from getDNSRecords until postDNSRecords
	d.dnsEntriesMu.Lock()
	defer d.dnsEntriesMu.Unlock()

	ctx := context.Background()

	zoneName := dns01.UnFqdn(authZone)

	domains, err := d.client.GetDomain(ctx, zoneName)
	if err != nil {
		return fmt.Errorf("versio: %w", err)
	}

	txtRecord := internal.Record{
		Type:  "TXT",
		Name:  info.EffectiveFQDN,
		Value: `"` + info.Value + `"`,
		TTL:   d.config.TTL,
	}

	// Add new txtRecord to existing array of DNSRecords.
	// We'll need all the dns_records to add a new TXT record.
	msg := &domains.DomainInfo
	msg.DNSRecords = append(msg.DNSRecords, txtRecord)

	_, err = d.client.UpdateDomain(ctx, zoneName, msg)
	if err != nil {
		return fmt.Errorf("versio: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("versio: could not find zone for domain %q: %w", domain, err)
	}

	// use mutex to prevent race condition from getDNSRecords until postDNSRecords
	d.dnsEntriesMu.Lock()
	defer d.dnsEntriesMu.Unlock()

	ctx := context.Background()

	zoneName := dns01.UnFqdn(authZone)

	domains, err := d.client.GetDomain(ctx, zoneName)
	if err != nil {
		return fmt.Errorf("versio: %w", err)
	}

	// loop through the existing entries and remove the specific record
	msg := &internal.DomainInfo{}

	for _, e := range domains.DomainInfo.DNSRecords {
		if e.Name != info.EffectiveFQDN {
			msg.DNSRecords = append(msg.DNSRecords, e)
		}
	}

	_, err = d.client.UpdateDomain(ctx, zoneName, msg)
	if err != nil {
		return fmt.Errorf("versio: %w", err)
	}

	return nil
}
