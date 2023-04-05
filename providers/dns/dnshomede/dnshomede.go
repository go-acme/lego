// Package dnshomede implements a DNS provider for solving the DNS-01 challenge using dnsHome.de.
package dnshomede

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/dnshomede/internal"
)

// Environment variables names.
const (
	envNamespace = "DNSHOMEDE_"

	EnvCredentials = envNamespace + "CREDENTIALS"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Credentials        map[string]string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 20*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, 2*time.Minute),
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

// NewDNSProvider returns a DNSProvider instance configured for dnsHome.de.
// Credentials must be passed in the environment variable: DNSHOMEDE_CREDENTIALS.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	values, err := env.Get(EnvCredentials)
	if err != nil {
		return nil, fmt.Errorf("dnshomede: %w", err)
	}

	credentials, err := parseCredentials(values[EnvCredentials])
	if err != nil {
		return nil, fmt.Errorf("dnshomede: %w", err)
	}

	config.Credentials = credentials

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnshomede: the configuration of the DNS provider is nil")
	}

	if len(config.Credentials) == 0 {
		return nil, errors.New("dnshomede: missing credentials")
	}

	for domain, password := range config.Credentials {
		if domain == "" {
			return nil, fmt.Errorf(`dnshomede: missing domain: "%s:%s"`, domain, password)
		}

		if password == "" {
			return nil, fmt.Errorf(`dnshomede: missing password: "%s:%s"`, domain, password)
		}
	}

	client := internal.NewClient(config.Credentials)

	return &DNSProvider{config: config, client: client}, nil
}

// Present updates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	err := d.client.Add(dns01.UnFqdn(info.EffectiveFQDN), info.Value)
	if err != nil {
		return fmt.Errorf("dnshomede: %w", err)
	}

	return nil
}

// CleanUp updates the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	err := d.client.Remove(dns01.UnFqdn(info.EffectiveFQDN), info.Value)
	if err != nil {
		return fmt.Errorf("dnshomede: %w", err)
	}

	return nil
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

func parseCredentials(raw string) (map[string]string, error) {
	credentials := make(map[string]string)

	credStrings := strings.Split(strings.TrimSuffix(raw, ","), ",")
	for _, credPair := range credStrings {
		data := strings.Split(credPair, ":")
		if len(data) != 2 {
			return nil, fmt.Errorf("invalid credential pair: %q", credPair)
		}

		credentials[strings.TrimSpace(data[0])] = strings.TrimSpace(data[1])
	}

	return credentials, nil
}
