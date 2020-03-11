package checkdomain

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
)

const (
	envEndpoint           = "CHECKDOMAIN_ENDPOINT"
	envToken              = "CHECKDOMAIN_TOKEN"
	envTTL                = "CHECKDOMAIN_TTL"
	envHTTPTimeout        = "CHECKDOMAIN_HTTP_TIMEOUT"
	envPropagationTimeout = "CHECKDOMAIN_PROPAGATION_TIMEOUT"
	envPollingInterval    = "CHECKDOMAIN_POLLING_INTERVAL"
)

const (
	defaultEndpoint = "https://api.checkdomain.de"
	defaultTTL      = 300
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Endpoint           *url.URL
	Token              string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(envTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(envPropagationTimeout, 5*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(envPollingInterval, 7*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(envHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements challenge.Provider for the checkdomain API
// specified at https://developer.checkdomain.de/reference/.
type DNSProvider struct {
	config *Config

	domainIDMu      sync.Mutex
	domainIDMapping map[string]int
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(envToken)
	if err != nil {
		return nil, fmt.Errorf("checkdomain: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[envToken]

	endpoint, err := url.Parse(env.GetOrDefaultString(envEndpoint, defaultEndpoint))
	if err != nil {
		return nil, fmt.Errorf("checkdomain: invalid %s: %w", envEndpoint, err)
	}
	config.Endpoint = endpoint

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config.Endpoint == nil {
		return nil, errors.New("checkdomain: invalid endpoint")
	}

	if config.Token == "" {
		return nil, errors.New("checkdomain: missing token")
	}

	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}

	return &DNSProvider{
		config:          config,
		domainIDMapping: make(map[string]int),
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	domainID, err := d.getDomainIDByName(domain)
	if err != nil {
		return fmt.Errorf("checkdomain: %w", err)
	}

	err = d.checkNameservers(domainID)
	if err != nil {
		return fmt.Errorf("checkdomain: %w", err)
	}

	name, value := dns01.GetRecord(domain, keyAuth)

	err = d.createRecord(domainID, &Record{
		Name:  name,
		TTL:   d.config.TTL,
		Type:  "TXT",
		Value: value,
	})

	if err != nil {
		return fmt.Errorf("checkdomain: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record previously created
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	domainID, err := d.getDomainIDByName(domain)
	if err != nil {
		return fmt.Errorf("checkdomain: %w", err)
	}

	err = d.checkNameservers(domainID)
	if err != nil {
		return fmt.Errorf("checkdomain: %w", err)
	}

	name, value := dns01.GetRecord(domain, keyAuth)

	err = d.deleteTXTRecord(domainID, name, value)
	if err != nil {
		return fmt.Errorf("checkdomain: %w", err)
	}

	d.domainIDMu.Lock()
	delete(d.domainIDMapping, name)
	d.domainIDMu.Unlock()

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
