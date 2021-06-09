// Package easydns implements a DNS provider for solving the DNS-01 challenge using EasyDNS API.
package easydns

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/miekg/dns"
)

// Environment variables names.
const (
	envNamespace = "EASYDNS_"

	EnvEndpoint = envNamespace + "ENDPOINT"
	EnvToken    = envNamespace + "TOKEN"
	EnvKey      = envNamespace + "KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Endpoint           *url.URL
	Token              string
	Key                string
	TTL                int
	HTTPClient         *http.Client
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
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
	config      *Config
	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	endpoint, err := url.Parse(env.GetOrDefaultString(EnvEndpoint, defaultEndpoint))
	if err != nil {
		return nil, fmt.Errorf("easydns: %w", err)
	}
	config.Endpoint = endpoint

	values, err := env.Get(EnvToken, EnvKey)
	if err != nil {
		return nil, fmt.Errorf("easydns: %w", err)
	}

	config.Token = values[EnvToken]
	config.Key = values[EnvKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for EasyDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("easydns: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("easydns: the API token is missing")
	}

	if config.Key == "" {
		return nil, errors.New("easydns: the API key is missing")
	}

	return &DNSProvider{config: config, recordIDs: map[string]string{}}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	apiHost, apiDomain := splitFqdn(fqdn)
	record := &zoneRecord{
		Domain: apiDomain,
		Host:   apiHost,
		Type:   "TXT",
		Rdata:  value,
		TTL:    strconv.Itoa(d.config.TTL),
		Prio:   "0",
	}

	recordID, err := d.addRecord(apiDomain, record)
	if err != nil {
		return fmt.Errorf("easydns: error adding zone record: %w", err)
	}

	key := getMapKey(fqdn, value)

	d.recordIDsMu.Lock()
	d.recordIDs[key] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, challenge := dns01.GetRecord(domain, keyAuth)

	key := getMapKey(fqdn, challenge)
	recordID, exists := d.recordIDs[key]
	if !exists {
		return nil
	}

	_, apiDomain := splitFqdn(fqdn)
	err := d.deleteRecord(apiDomain, recordID)

	d.recordIDsMu.Lock()
	defer delete(d.recordIDs, key)
	d.recordIDsMu.Unlock()

	if err != nil {
		return fmt.Errorf("easydns: %w", err)
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

func splitFqdn(fqdn string) (host, domain string) {
	parts := dns.SplitDomainName(fqdn)
	length := len(parts)

	host = strings.Join(parts[0:length-2], ".")
	domain = strings.Join(parts[length-2:length], ".")
	return
}

func getMapKey(fqdn, value string) string {
	return fqdn + "|" + value
}
