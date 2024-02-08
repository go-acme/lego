// Package alidns implements a DNS provider for solving the DNS-01 challenge using Alibaba Cloud DNS.
package alidns

import (
	"errors"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"golang.org/x/net/idna"
)

const defaultRegionID = "cn-hangzhou"

// Environment variables names.
const (
	envNamespace = "ALICLOUD_"

	EnvRAMRole       = envNamespace + "RAM_ROLE"
	EnvAccessKey     = envNamespace + "ACCESS_KEY"
	EnvSecretKey     = envNamespace + "SECRET_KEY"
	EnvSecurityToken = envNamespace + "SECURITY_TOKEN"
	EnvRegionID      = envNamespace + "REGION_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	RAMRole            string
	APIKey             string
	SecretKey          string
	SecurityToken      string
	RegionID           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *alidns.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Alibaba Cloud DNS.
// - If you're using the instance RAM role, the RAM role environment variable must be passed in: ALICLOUD_RAM_ROLE.
// - Other than that, credentials must be passed in the environment variables:
// ALICLOUD_ACCESS_KEY, ALICLOUD_SECRET_KEY, and optionally ALICLOUD_SECURITY_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.RegionID = env.GetOrFile(EnvRegionID)

	values, err := env.Get(EnvRAMRole)
	if err == nil {
		config.RAMRole = values[EnvRAMRole]
		return NewDNSProviderConfig(config)
	}

	values, err = env.Get(EnvAccessKey, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("alicloud: %w", err)
	}

	config.APIKey = values[EnvAccessKey]
	config.SecretKey = values[EnvSecretKey]
	config.SecurityToken = env.GetOrFile(EnvSecurityToken)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for alidns.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("alicloud: the configuration of the DNS provider is nil")
	}

	if config.RegionID == "" {
		config.RegionID = defaultRegionID
	}

	var credential auth.Credential
	switch {
	case config.RAMRole != "":
		credential = credentials.NewEcsRamRoleCredential(config.RAMRole)
	case config.APIKey != "" && config.SecretKey != "" && config.SecurityToken != "":
		credential = credentials.NewStsTokenCredential(config.APIKey, config.SecretKey, config.SecurityToken)
	case config.APIKey != "" && config.SecretKey != "":
		credential = credentials.NewAccessKeyCredential(config.APIKey, config.SecretKey)
	default:
		return nil, errors.New("alicloud: ram role or credentials missing")
	}

	conf := sdk.NewConfig().WithTimeout(config.HTTPTimeout)

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
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneName, err := d.getHostedZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("alicloud: %w", err)
	}

	recordAttributes, err := d.newTxtRecord(zoneName, info.EffectiveFQDN, info.Value)
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
	info := dns01.GetChallengeInfo(domain, keyAuth)

	records, err := d.findTxtRecords(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("alicloud: %w", err)
	}

	_, err = d.getHostedZone(info.EffectiveFQDN)
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

	var domains []alidns.DomainInDescribeDomains
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

	authZone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return "", fmt.Errorf("could not find zone for FQDN %q: %w", domain, err)
	}

	var hostedZone alidns.DomainInDescribeDomains
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

func (d *DNSProvider) findTxtRecords(fqdn string) ([]alidns.Record, error) {
	zoneName, err := d.getHostedZone(fqdn)
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
		if record.RR == recordName && record.Type == "TXT" {
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

	subDomain, err := dns01.ExtractSubDomain(fqdn, asciiDomain)
	if err != nil {
		return "", err
	}

	return subDomain, nil
}
