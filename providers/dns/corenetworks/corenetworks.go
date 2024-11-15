package corenetworks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/corenetworks/internal"
)

// Environment variables names.
const (
	envNamespace = "CORENETWORKS_"

	EnvLogin    = envNamespace + "LOGIN"
	EnvPassword = envNamespace + "PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Login              string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 3600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
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
}

// NewDNSProvider returns a DNSProvider instance configured for Core-Networks.
// Credentials must be passed in the environment variables: CORENETWORKS_LOGIN, CORENETWORKS_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvLogin, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("corenetworks: %w", err)
	}

	config := NewDefaultConfig()
	config.Login = values[EnvLogin]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Bluecat DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("corenetworks: the configuration of the DNS provider is nil")
	}

	if config.Login == "" || config.Password == "" {
		return nil, errors.New("corenetworks: credentials missing")
	}

	client := internal.NewClient(config.Login, config.Password)

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

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx, err := d.client.CreateAuthenticatedContext(context.Background())
	if err != nil {
		return fmt.Errorf("create authentication token: %w", err)
	}

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("corenetworks: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("corenetworks: %w", err)
	}

	record := internal.Record{
		Name: subDomain,
		TTL:  d.config.TTL,
		Type: "TXT",
		Data: info.Value,
	}

	err = d.client.AddRecord(ctx, dns01.UnFqdn(zone), record)
	if err != nil {
		return fmt.Errorf("corenetworks: add record: %w", err)
	}

	err = d.client.CommitRecords(ctx, dns01.UnFqdn(zone))
	if err != nil {
		return fmt.Errorf("corenetworks: commit records: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx, err := d.client.CreateAuthenticatedContext(context.Background())
	if err != nil {
		return fmt.Errorf("create authentication token: %w", err)
	}

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("corenetworks: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("corenetworks: %w", err)
	}

	record := internal.Record{
		Name: subDomain,
		TTL:  d.config.TTL,
		Type: "TXT",
		Data: info.Value,
	}

	err = d.client.DeleteRecords(ctx, dns01.UnFqdn(zone), record)
	if err != nil {
		return fmt.Errorf("corenetworks: delete records: %w", err)
	}

	err = d.client.CommitRecords(ctx, dns01.UnFqdn(zone))
	if err != nil {
		return fmt.Errorf("corenetworks: commit records: %w", err)
	}

	return nil
}
