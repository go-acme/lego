// Package vultr implements a DNS provider for solving the DNS-01 challenge using the Vultr DNS.
// See https://www.vultr.com/api/#dns
package vultr

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/vultr/govultr"
)

// Environment variables names.
const (
	envNamespace = "VULTR_"

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
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 30),
			// from Vultr Client
			Transport: &http.Transport{
				TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
			},
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *govultr.Client
}

// NewDNSProvider returns a DNSProvider instance with a configured Vultr client.
// Authentication uses the VULTR_API_KEY environment variable.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("vultr: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Vultr.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("vultr: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("vultr: credentials missing")
	}

	client := govultr.NewClient(config.HTTPClient, config.APIKey)

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the DNS-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zoneDomain, err := d.getHostedZone(ctx, domain)
	if err != nil {
		return fmt.Errorf("vultr: %w", err)
	}

	name := extractRecordName(fqdn, zoneDomain)

	err = d.client.DNSRecord.Create(ctx, zoneDomain, "TXT", name, `"`+value+`"`, d.config.TTL, 0)
	if err != nil {
		return fmt.Errorf("vultr: API call failed: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zoneDomain, records, err := d.findTxtRecords(ctx, domain, fqdn)
	if err != nil {
		return fmt.Errorf("vultr: %w", err)
	}

	var allErr []string
	for _, rec := range records {
		err := d.client.DNSRecord.Delete(ctx, zoneDomain, strconv.Itoa(rec.RecordID))
		if err != nil {
			allErr = append(allErr, err.Error())
		}
	}

	if len(allErr) > 0 {
		return errors.New(strings.Join(allErr, ": "))
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getHostedZone(ctx context.Context, domain string) (string, error) {
	domains, err := d.client.DNSDomain.List(ctx)
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	var hostedDomain govultr.DNSDomain
	for _, dom := range domains {
		if strings.HasSuffix(domain, dom.Domain) {
			if len(dom.Domain) > len(hostedDomain.Domain) {
				hostedDomain = dom
			}
		}
	}
	if hostedDomain.Domain == "" {
		return "", fmt.Errorf("no matching Vultr domain found for domain %s", domain)
	}

	return hostedDomain.Domain, nil
}

func (d *DNSProvider) findTxtRecords(ctx context.Context, domain, fqdn string) (string, []govultr.DNSRecord, error) {
	zoneDomain, err := d.getHostedZone(ctx, domain)
	if err != nil {
		return "", nil, err
	}

	var records []govultr.DNSRecord
	result, err := d.client.DNSRecord.List(ctx, zoneDomain)
	if err != nil {
		return "", records, fmt.Errorf("API call has failed: %w", err)
	}

	recordName := extractRecordName(fqdn, zoneDomain)
	for _, record := range result {
		if record.Type == "TXT" && record.Name == recordName {
			records = append(records, record)
		}
	}

	return zoneDomain, records, nil
}

func extractRecordName(fqdn, zone string) string {
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+zone); idx != -1 {
		return name[:idx]
	}
	return name
}
