// Package nearlyfreespeech implements a DNS provider for solving the DNS-01 challenge using nearlyfreespeech.net.
package nearlyfreespeech

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"golang.org/x/net/publicsuffix"
)

// Environment variables names.
const (
	envNamespace = "NFS_"

	EnvAPIKey = envNamespace + "API_KEY"
	EnvLogin  = envNamespace + "LOGIN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	Login              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	TTL                int
}

type challenge struct {
	domain   string
	key      string
	keyFqdn  string
	keyValue string
	tld      string
	sld      string
	host     string
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() (*Config, error) {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 3600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, dns01.DefaultPropagationTimeout),
	}, nil
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config

	rngMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for freemyip.com.
// Credentials must be passed in the environment variable: FREEMYIP_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	config, err := NewDefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("nearlyfreespeech: %w", err)
	}

	values, err := env.Get(EnvAPIKey, EnvLogin)
	if err != nil {
		return nil, fmt.Errorf("nearlyfreespeech: %w", err)
	}

	config.APIKey = values[EnvAPIKey]
	config.Login = values[EnvLogin]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for freemyip.com.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("nearlyfreespeech: the configuration of the DNS provider is nil")
	}

	if config.Login == "" && config.APIKey == "" {
		return nil, errors.New("nearlyfreespeech: API credentials are missing")
	}

	if config.APIKey == "" {
		return nil, errors.New("nearlyfreespeech: the API Login is missing")
	}

	if config.Login == "" {
		return nil, errors.New("nearlyfreespeech: the API Key is missing")
	}

	return &DNSProvider{config: config}, nil
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

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ch, err := newChallenge(domain, keyAuth)
	if err != nil {
		return fmt.Errorf("nearlyfreespeech: %w", err)
	}

	newRecord := TXTRecord{
		Name: ch.key,
		Type: "TXT",
		Data: ch.keyValue,
		TTL:  d.config.TTL,
	}

	clientConfig := ClientConfig{
		domain: domain,
	}

	err = d.AddRR(clientConfig, newRecord)
	if err != nil {
		return fmt.Errorf("nearlyfreespeech: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ch, err := newChallenge(domain, keyAuth)
	if err != nil {
		return fmt.Errorf("nearlyfreespeech: %w", err)
	}

	delRecord := TXTRecord{
		Name: ch.key,
		Type: "TXT",
		Data: ch.keyValue,
	}

	clientConfig := ClientConfig{
		domain: domain,
	}

	err = d.DeleteRR(clientConfig, delRecord)
	if err != nil {
		return fmt.Errorf("nearlyfreespeech: %w", err)
	}

	return nil
}

// newChallenge builds a challenge record from a domain name and a challenge authentication key.
func newChallenge(domain, keyAuth string) (*challenge, error) {
	domain = dns01.UnFqdn(domain)

	tld, _ := publicsuffix.PublicSuffix(domain)
	if tld == domain {
		return nil, fmt.Errorf("invalid domain name %q", domain)
	}

	parts := strings.Split(domain, ".")
	longest := len(parts) - strings.Count(tld, ".") - 1
	sld := parts[longest-1]

	var host string
	if longest >= 1 {
		host = strings.Join(parts[:longest-1], ".")
	}

	fqdn, value := dns01.GetRecord(domain, keyAuth)

	return &challenge{
		domain:   domain,
		key:      "_acme-challenge",
		keyFqdn:  fqdn,
		keyValue: value,
		tld:      tld,
		sld:      sld,
		host:     host,
	}, nil
}
