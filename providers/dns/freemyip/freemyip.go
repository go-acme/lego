// Package freemyip implements a DNS provider for solving the DNS-01 challenge using freemyip.com.
package freemyip

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/nrdcg/freemyip"
)

// Environment variables names.
const (
	envNamespace = "FREEMYIP_"

	EnvToken = envNamespace + "TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token              string
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
	client *freemyip.Client
}

// NewDNSProvider returns a DNSProvider instance configured for freemyip.com.
// Credentials must be passed in the environment variable: FREEMYIP_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken)
	if err != nil {
		return nil, fmt.Errorf("freemyip: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for freemyip.com.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("freemyip: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("freemyip: missing credentials")
	}

	client := freemyip.New(config.Token, true)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
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

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, freemyip.RootDomain)
	if err != nil {
		return fmt.Errorf("freemyip: %w", err)
	}

	_, err = d.client.EditTXTRecord(context.Background(), subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("freemyip: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, freemyip.RootDomain)
	if err != nil {
		return fmt.Errorf("freemyip: %w", err)
	}

	_, err = d.client.DeleteTXTRecord(context.Background(), subDomain)
	if err != nil {
		return fmt.Errorf("freemyip: %w", err)
	}

	return nil
}
