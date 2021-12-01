// Package dnspod implements a DNS provider for solving the DNS-01 challenge using dnspod DNS.
package dnspod

import (
	"errors"
	"fmt"
	"golang.org/x/net/idna"
	"math"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

// Environment variables names.
const (
	envNamespace = "DNSPOD_"

	EnvSecretID     = envNamespace + "SECRET_ID"
	EnvSecretKey    = envNamespace + "SECRET_KEY"
	EnvSessionToken = envNamespace + "SESSION_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	SecretID           string
	SecretKey          string
	SessionToken       string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
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
	values, err := env.Get(EnvSecretID, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("dnspod: %w", err)
	}

	config := NewDefaultConfig()
	config.SecretID = values[EnvSecretID]
	config.SecretKey = values[EnvSecretKey]
	config.SessionToken = env.GetOrFile(EnvSessionToken)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for dnspod.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnspod: the configuration of the DNS provider is nil")
	}

	var credential *common.Credential

	switch {
	case config.SecretID != "" && config.SecretKey != "" && config.SessionToken != "":
		credential = common.NewTokenCredential(config.SecretID, config.SecretKey, config.SessionToken)
	case config.SecretID != "" && config.SecretKey != "":
		credential = common.NewCredential(config.SecretID, config.SecretKey)
	default:
		return nil, fmt.Errorf("dnspod: credentials missing")
	}

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	cpf.HttpProfile.ReqTimeout = int(math.Round(config.HTTPTimeout.Seconds()))
	client, err := dnspod.NewClient(credential, "", cpf)
	if err != nil {
		return nil, fmt.Errorf("dnspod: credentials failed: %w", err)
	}

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	zoneID, zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return err
	}
	recordName, err := extractRecordName(fqdn, zoneName)
	if err != nil {
		return err
	}
	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(zoneName)
	request.DomainId = common.Uint64Ptr(zoneID)
	request.SubDomain = common.StringPtr(recordName)
	request.RecordType = common.StringPtr("TXT")
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(value)
	request.TTL = common.Uint64Ptr(uint64(d.config.TTL))

	_, err = d.client.CreateRecord(request)
	if err != nil {
		return fmt.Errorf("dnspod: API call failed: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	zoneID, zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return err
	}

	records, err := d.findTxtRecords(zoneID, zoneName, fqdn)
	if err != nil {
		return err
	}

	for _, record := range records {
		request := dnspod.NewDeleteRecordRequest()
		request.Domain = common.StringPtr(zoneName)
		request.DomainId = common.Uint64Ptr(zoneID)
		request.RecordId = record.RecordId

		_, err := d.client.DeleteRecord(request)
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

func (d *DNSProvider) getHostedZone(domain string) (uint64, string, error) {
	request := dnspod.NewDescribeDomainListRequest()

	var domains []*dnspod.DomainListItem

	for {
		request.Offset = common.Int64Ptr(int64(len(domains)))
		response, err := d.client.DescribeDomainList(request)
		if err != nil {
			return 0, "", fmt.Errorf("dnspod: API call failed: %w", err)
		}

		domains = append(domains, response.Response.DomainList...)

		if uint64(len(domains)) >= *response.Response.DomainCountInfo.AllTotal {
			break
		}
	}

	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return 0, "", err
	}

	var hostedZone *dnspod.DomainListItem
	for _, zone := range domains {
		if *zone.Name == dns01.UnFqdn(authZone) {
			hostedZone = zone
		}
	}

	if hostedZone == nil {
		return 0, "", fmt.Errorf("zone %s not found in dnspod for domain %s", authZone, domain)
	}

	return *hostedZone.DomainId, *hostedZone.Name, nil
}

func (d *DNSProvider) findTxtRecords(zoneID uint64, zoneName, fqdn string) ([]*dnspod.RecordListItem, error) {
	recordName, err := extractRecordName(fqdn, zoneName)
	if err != nil {
		return nil, err
	}
	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(zoneName)
	request.DomainId = common.Uint64Ptr(zoneID)
	request.Subdomain = common.StringPtr(recordName)
	request.RecordType = common.StringPtr("TXT")
	request.RecordLine = common.StringPtr("默认")
	response, err := d.client.DescribeRecordList(request)
	if err != nil {
		return nil, err
	}
	return response.Response.RecordList, nil
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
