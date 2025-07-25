// Package alidns implements a DNS provider for solving the DNS-01 challenge using Alibaba Cloud DNS.
package alidns

import (
	"context"
	"errors"
	"fmt"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/aliyun/credentials-go/credentials"
	alidns "github.com/go-acme/alidns-20150109/v4/client"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/ptr"
	"golang.org/x/net/idna"
)

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

const defaultRegionID = "cn-hangzhou"

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

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

	cfg := new(openapi.Config).
		SetRegionId(config.RegionID).
		SetReadTimeout(int(config.HTTPTimeout.Milliseconds()))

	switch {
	case config.RAMRole != "":
		// https://www.alibabacloud.com/help/en/ecs/user-guide/attach-an-instance-ram-role-to-an-ecs-instance
		credentialsCfg := new(credentials.Config).
			SetType("ecs_ram_role").
			SetRoleName(config.RAMRole)

		credentialClient, err := credentials.NewCredential(credentialsCfg)
		if err != nil {
			return nil, fmt.Errorf("alicloud: new credential: %w", err)
		}

		cfg = cfg.SetCredential(credentialClient)

	case config.APIKey != "" && config.SecretKey != "" && config.SecurityToken != "":
		cfg = cfg.
			SetAccessKeyId(config.APIKey).
			SetAccessKeySecret(config.SecretKey).
			SetSecurityToken(config.SecurityToken)

	case config.APIKey != "" && config.SecretKey != "":
		cfg = cfg.
			SetAccessKeyId(config.APIKey).
			SetAccessKeySecret(config.SecretKey)

	default:
		return nil, errors.New("alicloud: ram role or credentials missing")
	}

	client, err := alidns.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("alicloud: new client: %w", err)
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

	recordRequest, err := d.newTxtRecord(zoneName, info.EffectiveFQDN, info.Value)
	if err != nil {
		return err
	}

	_, err = alidns.AddDomainRecordWithContext(context.Background(), d.client, recordRequest, &dara.RuntimeOptions{})
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
		request := &alidns.DeleteDomainRecordRequest{
			RecordId: rec.RecordId,
		}

		_, err = alidns.DeleteDomainRecordWithContext(context.Background(), d.client, request, &dara.RuntimeOptions{})
		if err != nil {
			return fmt.Errorf("alicloud: %w", err)
		}
	}

	return nil
}

func (d *DNSProvider) getHostedZone(domain string) (string, error) {
	request := new(alidns.DescribeDomainsRequest)

	var domains []*alidns.DescribeDomainsResponseBodyDomainsDomain

	var startPage int64 = 1

	for {
		request.SetPageNumber(startPage)

		response, err := alidns.DescribeDomainsWithContext(context.Background(), d.client, request, &dara.RuntimeOptions{})
		if err != nil {
			return "", fmt.Errorf("API call failed: %w", err)
		}

		domains = append(domains, response.Body.Domains.Domain...)

		if ptr.Deref(response.Body.PageNumber)*ptr.Deref(response.Body.PageSize) >= ptr.Deref(response.Body.TotalCount) {
			break
		}

		startPage++
	}

	authZone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return "", fmt.Errorf("could not find zone: %w", err)
	}

	var hostedZone *alidns.DescribeDomainsResponseBodyDomainsDomain
	for _, zone := range domains {
		if ptr.Deref(zone.DomainName) == dns01.UnFqdn(authZone) || ptr.Deref(zone.PunyCode) == dns01.UnFqdn(authZone) {
			hostedZone = zone
		}
	}

	if hostedZone == nil || ptr.Deref(hostedZone.DomainId) == "" {
		return "", fmt.Errorf("zone %s not found in AliDNS for domain %s", authZone, domain)
	}

	return ptr.Deref(hostedZone.DomainName), nil
}

func (d *DNSProvider) newTxtRecord(zone, fqdn, value string) (*alidns.AddDomainRecordRequest, error) {
	rr, err := extractRecordName(fqdn, zone)
	if err != nil {
		return nil, err
	}

	return new(alidns.AddDomainRecordRequest).
		SetType("TXT").
		SetDomainName(zone).
		SetRR(rr).
		SetValue(value).
		SetTTL(int64(d.config.TTL)), nil
}

func (d *DNSProvider) findTxtRecords(fqdn string) ([]*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, error) {
	zoneName, err := d.getHostedZone(fqdn)
	if err != nil {
		return nil, err
	}

	request := new(alidns.DescribeDomainRecordsRequest).
		SetDomainName(zoneName).
		SetPageSize(500)

	var records []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord

	result, err := alidns.DescribeDomainRecordsWithContext(context.Background(), d.client, request, &dara.RuntimeOptions{})
	if err != nil {
		return records, fmt.Errorf("API call has failed: %w", err)
	}

	recordName, err := extractRecordName(fqdn, zoneName)
	if err != nil {
		return nil, err
	}

	for _, record := range result.Body.DomainRecords.Record {
		if ptr.Deref(record.RR) == recordName && ptr.Deref(record.Type) == "TXT" {
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
