package hurricane

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/hurricane/internal"
)

// Environment variables names.
const (
	envNamespace = "HURRICANE_"

	EnvTokens = envNamespace + "TOKENS"

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
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 300*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for Hurricane Electric.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	values, err := env.Get(EnvTokens)
	if err != nil {
		return nil, fmt.Errorf("hurricane: %w", err)
	}

	// parse tokens and create map
	credStrings := strings.Split(values[EnvTokens], ",")
	creds := make(map[string]string)
	for _, credPair := range credStrings {
		data := strings.Split(credPair, ":")
		if len(data) != 2 {
			return nil, fmt.Errorf("hurricane: Read incorrect credential pair %s", credPair)
		}
		creds[data[0]] = data[1]
	}
	config.Credentials = creds

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("hurricane: the configuration of the DNS provider is nil")
	}

	if config.Credentials == nil {
		return nil, errors.New("hurricane: credentials missing")
	}

	client := internal.NewClient(config.Credentials)

	return &DNSProvider{config: config, client: client}, nil
}

// Present updates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	_, txtRecord := dns01.GetRecord(domain, keyAuth)

	err := d.client.UpdateTxtRecord(domain, txtRecord)
	if err != nil {
		return fmt.Errorf("hurricane: %w", err)
	}

	return nil
}

// CleanUp updates the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, _, _ string) error {
	err := d.client.UpdateTxtRecord(domain, ".")
	if err != nil {
		return fmt.Errorf("hurricane: %w", err)
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
