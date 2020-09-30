// Package auroradns implements a DNS provider for solving the DNS-01 challenge using Aurora DNS.
package auroradns

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/nrdcg/auroradns"
)

const defaultBaseURL = "https://api.auroradns.eu"

// Environment variables names.
const (
	envNamespace = "AURORA_"

	EnvUserID   = envNamespace + "USER_ID"
	EnvKey      = envNamespace + "KEY"
	EnvEndpoint = envNamespace + "ENDPOINT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	UserID             string
	Key                string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	recordIDs   map[string]string
	recordIDsMu sync.Mutex
	config      *Config
	client      *auroradns.Client
}

// NewDNSProvider returns a DNSProvider instance configured for AuroraDNS.
// Credentials must be passed in the environment variables:
// AURORA_USER_ID and AURORA_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUserID, EnvKey)
	if err != nil {
		return nil, fmt.Errorf("aurora: %w", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = env.GetOrFile(EnvEndpoint)
	config.UserID = values[EnvUserID]
	config.Key = values[EnvKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for AuroraDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("aurora: the configuration of the DNS provider is nil")
	}

	if config.UserID == "" || config.Key == "" {
		return nil, errors.New("aurora: some credentials information are missing")
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	tr, err := auroradns.NewTokenTransport(config.UserID, config.Key)
	if err != nil {
		return nil, fmt.Errorf("aurora: %w", err)
	}

	client, err := auroradns.NewClient(tr.Client(), auroradns.WithBaseURL(config.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("aurora: %w", err)
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]string),
	}, nil
}

// Present creates a TXT record to fulfill the DNS-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.CreateRecord(domain, token, fqdn, value)
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) CreateRecord(domain, token, fqdn, value string) error {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("aurora: could not determine zone for domain %q: %w", domain, err)
	}

	// 1. Aurora will happily create the TXT record when it is provided a fqdn,
	//    but it will only appear in the control panel and will not be
	//    propagated to DNS servers. Extract and use subdomain instead.
	// 2. A trailing dot in the fqdn will cause Aurora to add a trailing dot to
	//    the subdomain, resulting in _acme-challenge..<domain> rather
	//    than _acme-challenge.<domain>

	subdomain := fqdn[0 : len(fqdn)-len(authZone)-1]

	authZone = dns01.UnFqdn(authZone)

	zone, err := d.getZoneInformationByName(authZone)
	if err != nil {
		return fmt.Errorf("aurora: could not create record: %w", err)
	}

	record := auroradns.Record{
		RecordType: "TXT",
		Name:       subdomain,
		Content:    value,
		TTL:        d.config.TTL,
	}

	newRecord, _, err := d.client.CreateRecord(zone.ID, record)
	if err != nil {
		return fmt.Errorf("aurora: could not create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.DeleteRecord(domain, token, fqdn, value)
}

// DeleteRecord removes a creates a TXT record from the provider.
func (d *DNSProvider) DeleteRecord(domain, token, fqdn, value string) error {
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("unknown recordID for %q", fqdn)
	}

	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("could not determine zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	zone, err := d.getZoneInformationByName(authZone)
	if err != nil {
		return err
	}

	_, _, err = d.client.DeleteRecord(zone.ID, recordID)
	if err != nil {
		return err
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getZoneInformationByName(name string) (auroradns.Zone, error) {
	zs, _, err := d.client.ListZones()
	if err != nil {
		return auroradns.Zone{}, err
	}

	for _, element := range zs {
		if element.Name == name {
			return element, nil
		}
	}

	return auroradns.Zone{}, errors.New("could not find Zone record")
}
