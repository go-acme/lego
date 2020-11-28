// Package vultr implements a DNS provider for solving the DNS-01 challenge using the Vultr DNS.
// See https://www.vultr.com/api/#dns
package vultr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/vultr/govultr/v2"
	"golang.org/x/oauth2"
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
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *govultr.Client
}

// NewDNSProvider returns a DNSProvider instance with a configured Vultr client.
// Authentication uses the VULTR_API_KEY environment variable.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("vultr: %w", err)
	}

	config := NewDefaultConfig()
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

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: config.HTTPTimeout,
			Transport: &oauth2.Transport{
				Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.APIKey}),
			},
		}
	}

	client := govultr.NewClient(httpClient)

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

	req := govultr.DomainRecordReq{
		Name:     name,
		Type:     "TXT",
		Data:     `"` + value + `"`,
		TTL:      d.config.TTL,
		Priority: func(v int) *int { return &v }(0),
	}
	_, err = d.client.DomainRecord.Create(ctx, zoneDomain, &req)
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
		err := d.client.DomainRecord.Delete(ctx, zoneDomain, rec.ID)
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
	listOptions := &govultr.ListOptions{PerPage: 25}

	var hostedDomain govultr.Domain

	for {
		domains, meta, err := d.client.Domain.List(ctx, listOptions)
		if err != nil {
			return "", fmt.Errorf("API call failed: %w", err)
		}

		for _, dom := range domains {
			if strings.HasSuffix(domain, dom.Domain) && len(dom.Domain) > len(hostedDomain.Domain) {
				hostedDomain = dom
			}
		}

		if domain == hostedDomain.Domain {
			break
		}

		if meta.Links.Next == "" {
			break
		}

		listOptions.Cursor = meta.Links.Next
	}

	if hostedDomain.Domain == "" {
		return "", fmt.Errorf("no matching domain found for domain %s", domain)
	}

	return hostedDomain.Domain, nil
}

func (d *DNSProvider) findTxtRecords(ctx context.Context, domain, fqdn string) (string, []govultr.DomainRecord, error) {
	zoneDomain, err := d.getHostedZone(ctx, domain)
	if err != nil {
		return "", nil, err
	}

	listOptions := &govultr.ListOptions{PerPage: 25}

	var records []govultr.DomainRecord
	for {
		result, meta, err := d.client.DomainRecord.List(ctx, zoneDomain, listOptions)
		if err != nil {
			return "", records, fmt.Errorf("API call has failed: %w", err)
		}

		recordName := extractRecordName(fqdn, zoneDomain)
		for _, record := range result {
			if record.Type == "TXT" && record.Name == recordName {
				records = append(records, record)
			}
		}

		if meta.Links.Next == "" {
			break
		}

		listOptions.Cursor = meta.Links.Next
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
