// Package yandex implements a DNS provider for solving the DNS-01 challenge using Yandex.
package yandex

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/providers/dns/yandex/internal"
	"github.com/miekg/dns"
)

const defaultTTL = 21600

// Environment variables names.
const (
	envNamespace = "YANDEX_"

	EnvPddToken = envNamespace + "PDD_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	PddToken           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *internal.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Yandex.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvPddToken)
	if err != nil {
		return nil, fmt.Errorf("yandex: %v", err)
	}

	config := NewDefaultConfig(conf)
	config.PddToken = values[EnvPddToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Yandex.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("yandex: the configuration of the DNS provider is nil")
	}

	if config.PddToken == "" {
		return nil, fmt.Errorf("yandex: credentials missing")
	}

	client, err := internal.NewClient(config.PddToken)
	if err != nil {
		return nil, fmt.Errorf("yandex: %v", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	rootDomain, subDomain, err := splitDomain(fqdn)
	if err != nil {
		return fmt.Errorf("yandex: %v", err)
	}

	data := internal.Record{
		Domain:    rootDomain,
		SubDomain: subDomain,
		Type:      "TXT",
		TTL:       d.config.TTL,
		Content:   value,
	}

	_, err = d.client.AddRecord(data)
	if err != nil {
		return fmt.Errorf("yandex: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	rootDomain, subDomain, err := splitDomain(fqdn)
	if err != nil {
		return fmt.Errorf("yandex: %v", err)
	}

	records, err := d.client.GetRecords(rootDomain)
	if err != nil {
		return fmt.Errorf("yandex: %v", err)
	}

	var record *internal.Record
	for _, rcd := range records {
		rcd := rcd
		if rcd.Type == "TXT" && rcd.SubDomain == subDomain && rcd.Content == value {
			record = &rcd
			break
		}
	}

	if record == nil {
		return fmt.Errorf("yandex: TXT record not found for domain: %s", domain)
	}

	data := internal.Record{
		ID:     record.ID,
		Domain: rootDomain,
	}

	_, err = d.client.RemoveRecord(data)
	if err != nil {
		return fmt.Errorf("yandex: %v", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func splitDomain(full string) (string, string, error) {
	split := dns.Split(full)
	if len(split) < 2 {
		return "", "", fmt.Errorf("unsupported domain: %s", full)
	}

	if len(split) == 2 {
		return full, "", nil
	}

	domain := full[split[len(split)-2]:]
	subDomain := full[:split[len(split)-2]-1]

	return domain, subDomain, nil
}
