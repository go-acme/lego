// Package alidns implements a DNS provider for solving the DNS-01 challenge using Alibaba Cloud DNS.
package alidns

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"golang.org/x/net/idna"
)

const defaultRegionID = "cn-hangzhou"

// Environment variables names.
const (
	envNamespace = "ALICLOUD_"

	EnvAccessKey = envNamespace + "ACCESS_KEY"
	EnvSecretKey = envNamespace + "SECRET_KEY"
	EnvRegionID  = envNamespace + "REGION_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	SecretKey          string
	RegionID           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 10*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *alidns.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Alibaba Cloud DNS.
// Credentials must be passed in the environment variables:
// ALICLOUD_ACCESS_KEY and ALICLOUD_SECRET_KEY.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvAccessKey, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("alicloud: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.APIKey = values[EnvAccessKey]
	config.SecretKey = values[EnvSecretKey]
	config.RegionID = env.GetOrFile(conf, EnvRegionID)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for alidns.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("alicloud: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("alicloud: credentials missing")
	}

	if len(config.RegionID) == 0 {
		config.RegionID = defaultRegionID
	}

	conf := sdk.NewConfig().WithTimeout(config.HTTPTimeout)
	credential := credentials.NewAccessKeyCredential(config.APIKey, config.SecretKey)

	client, err := alidns.NewClientWithOptions(config.RegionID, conf, credential)
	if err != nil {
		return nil, fmt.Errorf("alicloud: credentials failed: %w", err)
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("alicloud: %w", err)
	}

	recordAttributes, err := d.newTxtRecord(zoneName, fqdn, value)
	if err != nil {
		return err
	}

	_, err = d.client.AddDomainRecord(recordAttributes)
	if err != nil {
		return fmt.Errorf("alicloud: API call failed: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	records, err := d.findTxtRecords(domain, fqdn)
	if err != nil {
		return fmt.Errorf("alicloud: %w", err)
	}

	_, err = d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("alicloud: %w", err)
	}

	for _, rec := range records {
		request := alidns.CreateDeleteDomainRecordRequest()
		request.RecordId = rec.RecordId
		_, err = d.client.DeleteDomainRecord(request)
		if err != nil {
			return fmt.Errorf("alicloud: %w", err)
		}
	}
	return nil
}

func (d *DNSProvider) getHostedZone(domain string) (string, error) {
	request := alidns.CreateDescribeDomainsRequest()

	var domains []alidns.Domain
	startPage := 1

	for {
		request.PageNumber = requests.NewInteger(startPage)

		response, err := d.client.DescribeDomains(request)
		if err != nil {
			return "", fmt.Errorf("API call failed: %w", err)
		}

		domains = append(domains, response.Domains.Domain...)

		if response.PageNumber*response.PageSize >= response.TotalCount {
			break
		}

		startPage++
	}

	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return "", err
	}

	var hostedZone alidns.Domain
	for _, zone := range domains {
		if zone.DomainName == dns01.UnFqdn(authZone) || zone.PunyCode == dns01.UnFqdn(authZone) {
			hostedZone = zone
		}
	}

	if hostedZone.DomainId == "" {
		return "", fmt.Errorf("zone %s not found in AliDNS for domain %s", authZone, domain)
	}

	return hostedZone.DomainName, nil
}

func (d *DNSProvider) newTxtRecord(zone, fqdn, value string) (*alidns.AddDomainRecordRequest, error) {
	request := alidns.CreateAddDomainRecordRequest()
	request.Type = "TXT"
	request.DomainName = zone

	var err error
	request.RR, err = extractRecordName(fqdn, zone)
	if err != nil {
		return nil, err
	}

	request.Value = value
	request.TTL = requests.NewInteger(d.config.TTL)

	return request, nil
}

func (d *DNSProvider) findTxtRecords(domain, fqdn string) ([]alidns.Record, error) {
	zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return nil, err
	}

	request := alidns.CreateDescribeDomainRecordsRequest()
	request.DomainName = zoneName
	request.PageSize = requests.NewInteger(500)

	var records []alidns.Record

	result, err := d.client.DescribeDomainRecords(request)
	if err != nil {
		return records, fmt.Errorf("API call has failed: %w", err)
	}

	recordName, err := extractRecordName(fqdn, zoneName)
	if err != nil {
		return nil, err
	}

	for _, record := range result.DomainRecords.Record {
		if record.RR == recordName {
			records = append(records, record)
		}
	}
	return records, nil
}

func extractRecordName(fqdn, zone string) (string, error) {
	asciiDomain, err := idna.ToASCII(zone)
	if err != nil {
		return "", fmt.Errorf("fail to convert punycode: %w", err)
	}

	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+asciiDomain); idx != -1 {
		return name[:idx], nil
	}
	return name, nil
}
