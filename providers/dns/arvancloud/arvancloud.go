// Package arvancloud implements a DNS provider for solving the DNS-01 challenge using ArvanCloud DNS.
package arvancloud

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/providers/dns/arvancloud/internal"
)

const minTTL = 600

// Environment variables names.
const (
	envNamespace = "ARVANCLOUD_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for ArvanCloud.
// Credentials must be passed in the environment variable: ARVANCLOUD_API_KEY.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("arvancloud: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ArvanCloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("arvancloud: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("arvancloud: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("arvancloud: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := internal.NewClient(config.APIKey)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]string),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := getZone(fqdn)
	if err != nil {
		return err
	}

	record := internal.DNSRecord{
		Type:          "txt",
		Name:          extractRecordName(fqdn, authZone),
		Value:         internal.TXTRecordValue{Text: value},
		TTL:           d.config.TTL,
		UpstreamHTTPS: "default",
		IPFilterMode: &internal.IPFilterMode{
			Count:     "single",
			GeoFilter: "none",
			Order:     "none",
		},
	}

	newRecord, err := d.client.CreateRecord(authZone, record)
	if err != nil {
		return fmt.Errorf("arvancloud: failed to add TXT record: fqdn=%s: %w", fqdn, err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	authZone, err := getZone(fqdn)
	if err != nil {
		return err
	}

	// gets the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("arvancloud: unknown record ID for '%s' '%s'", fqdn, token)
	}

	if err := d.client.DeleteRecord(authZone, recordID); err != nil {
		return fmt.Errorf("arvancloud: failed to delate TXT record: id=%s: %w", recordID, err)
	}

	// deletes record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

func getZone(fqdn string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", err
	}

	return dns01.UnFqdn(authZone), nil
}

func extractRecordName(fqdn, zone string) string {
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+zone); idx != -1 {
		return name[:idx]
	}
	return name
}
