// Package dnspod implements a DNS provider for solving the DNS-01 challenge using dnspod DNS.
package dnspod

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/nrdcg/dnspod-go"
)

// Environment variables names.
const (
	envNamespace = "DNSPOD_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	LoginToken         string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *dnspod.Client
}

// NewDNSProvider returns a DNSProvider instance configured for dnspod.
// Credentials must be passed in the environment variables: DNSPOD_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("dnspod: %w", err)
	}

	config := NewDefaultConfig()
	config.LoginToken = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for dnspod.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnspod: the configuration of the DNS provider is nil")
	}

	if config.LoginToken == "" {
		return nil, errors.New("dnspod: credentials missing")
	}

	params := dnspod.CommonParams{LoginToken: config.LoginToken, Format: "json"}

	client := dnspod.NewClient(params)
	client.HTTPClient = config.HTTPClient

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneID, zoneName, err := d.getHostedZone(info.EffectiveFQDN)
	if err != nil {
		return err
	}

	recordAttributes, err := d.newTxtRecord(zoneName, info.EffectiveFQDN, info.Value, d.config.TTL)
	if err != nil {
		return err
	}

	_, _, err = d.client.Records.Create(zoneID, *recordAttributes)
	if err != nil {
		return fmt.Errorf("API call failed: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneID, zoneName, err := d.getHostedZone(info.EffectiveFQDN)
	if err != nil {
		return err
	}

	records, err := d.findTxtRecords(info.EffectiveFQDN, zoneID, zoneName)
	if err != nil {
		return err
	}

	for _, rec := range records {
		_, err := d.client.Records.Delete(zoneID, rec.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getHostedZone(domain string) (string, string, error) {
	zones, _, err := d.client.Domains.List()
	if err != nil {
		return "", "", fmt.Errorf("API call failed: %w", err)
	}

	authZone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return "", "", fmt.Errorf("could not find zone for FQDN %q: %w", domain, err)
	}

	var hostedZone dnspod.Domain
	for _, zone := range zones {
		if zone.Name == dns01.UnFqdn(authZone) {
			hostedZone = zone
		}
	}

	if hostedZone.ID == "" || hostedZone.ID == "0" {
		return "", "", fmt.Errorf("zone %s not found in dnspod for domain %s", authZone, domain)
	}

	return hostedZone.ID.String(), hostedZone.Name, nil
}

func (d *DNSProvider) newTxtRecord(zone, fqdn, value string, ttl int) (*dnspod.Record, error) {
	subDomain, err := dns01.ExtractSubDomain(fqdn, zone)
	if err != nil {
		return nil, err
	}

	return &dnspod.Record{
		Type:  "TXT",
		Name:  subDomain,
		Value: value,
		Line:  "默认",
		TTL:   strconv.Itoa(ttl),
	}, nil
}

func (d *DNSProvider) findTxtRecords(fqdn, zoneID, zoneName string) ([]dnspod.Record, error) {
	subDomain, err := dns01.ExtractSubDomain(fqdn, zoneName)
	if err != nil {
		return nil, err
	}

	var records []dnspod.Record
	result, _, err := d.client.Records.List(zoneID, subDomain)
	if err != nil {
		return records, fmt.Errorf("API call has failed: %w", err)
	}

	for _, record := range result {
		if record.Name == subDomain {
			records = append(records, record)
		}
	}

	return records, nil
}
