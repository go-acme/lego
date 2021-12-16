package safedns

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables.
const (
	envNamespace = "SAFEDNS_"

	EnvAuthToken          = envNamespace + "AUTH_TOKEN"
	EnvTTL                = envNamespace + "TTL"
	EnvAPITimeout         = envNamespace + "API_TIMEOUT"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

type Config struct {
	BaseURL            string
	AuthToken          string
	TTL                int
	APITimeout         time.Duration
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		TTL:                env.GetOrDefaultInt(EnvTTL, 30),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 60*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 5*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvAPITimeout, 30*time.Second),
		},
	}
}

type DNSProvider struct {
	config      *Config
	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAuthToken)
	if err != nil {
		return nil, fmt.Errorf("safedns: %w", err)
	}

	config := NewDefaultConfig()
	config.AuthToken = values[EnvAuthToken]
	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("safedns: supplied configuration was nil")
	}

	if config.AuthToken == "" {
		return nil, errors.New("safedns: credentials missing")
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &DNSProvider{
		config:    config,
		recordIDs: make(map[string]int),
	}, nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	respData, err := d.addTxtRecord(fqdn, value)
	if err != nil {
		return fmt.Errorf("safedns: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = respData.Data.ID
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("safedns: %w", err)
	}

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("safedns: unknown record ID for '%s'", fqdn)
	}

	err = d.removeTxtRecord(authZone, recordID)
	if err != nil {
		return fmt.Errorf("safedns: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}
