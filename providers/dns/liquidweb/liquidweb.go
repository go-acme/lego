// Package liquidweb implements a DNS provider for solving the DNS-01 challenge using Liquid Web.
package liquidweb

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	lw "github.com/liquidweb/liquidweb-go/client"
	"github.com/liquidweb/liquidweb-go/network"
)

const defaultBaseURL = "https://api.stormondemand.com"

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	Username           string
	Password           string
	Zone               string
	TTL                int
	PollingInterval    time.Duration
	PropagationTimeout time.Duration
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	config := &Config{
		BaseURL:            defaultBaseURL,
		TTL:                env.GetOrDefaultInt("LIQUID_WEB_TTL", 300),
		PollingInterval:    env.GetOrDefaultSecond("LIQUID_WEB_POLLING_INTERVAL", 2*time.Second),
		PropagationTimeout: env.GetOrDefaultSecond("LIQUID_WEB_PROPAGATION_TIMEOUT", 2*time.Minute),
		HTTPTimeout:        env.GetOrDefaultSecond("LIQUID_WEB_HTTP_TIMEOUT", 1*time.Minute),
	}

	return config
}

// DNSProvider is an implementation of the challenge.Provider interface
// that uses Liquid Web's REST API to manage TXT records for a domain.
type DNSProvider struct {
	config      *Config
	client      *lw.API
	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Liquid Web.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("LIQUID_WEB_USERNAME", "LIQUID_WEB_PASSWORD", "LIQUID_WEB_ZONE")
	if err != nil {
		return nil, fmt.Errorf("liquidweb: %v", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = env.GetOrFile("LIQUID_WEB_URL")
	config.Username = values["LIQUID_WEB_USERNAME"]
	config.Password = values["LIQUID_WEB_PASSWORD"]
	config.Zone = values["LIQUID_WEB_ZONE"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Liquid Web.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("liquidweb: the configuration of the DNS provider is nil")
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	if config.Zone == "" {
		return nil, fmt.Errorf("liquidweb: zone is missing")
	}

	if config.Username == "" {
		return nil, fmt.Errorf("liquidweb: username is missing")
	}

	if config.Password == "" {
		return nil, fmt.Errorf("liquidweb: password is missing")
	}

	// Initialize LW client.
	client, err := lw.NewAPI(config.Username, config.Password, config.BaseURL, int(config.HTTPTimeout.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("liquidweb: could not create Liquid Web API client: %v", err)
	}

	return &DNSProvider{
		config:    config,
		recordIDs: make(map[string]int),
		client:    client,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (time.Duration, time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	params := &network.DNSRecordParams{
		Name:  dns01.UnFqdn(fqdn),
		RData: strconv.Quote(value),
		Type:  "TXT",
		Zone:  d.config.Zone,
		TTL:   d.config.TTL,
	}

	dnsEntry, err := d.client.NetworkDNS.Create(params)
	if err != nil {
		return fmt.Errorf("liquidweb: could not create TXT record: %v", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = int(dnsEntry.ID)
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("liquidweb: unknown record ID for '%s'", domain)
	}

	params := &network.DNSRecordParams{ID: recordID}
	_, err := d.client.NetworkDNS.Delete(params)
	if err != nil {
		return fmt.Errorf("liquidweb: could not remove TXT record: %v", err)
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}
