package tencentcloud

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

// Environment variables names.
const (
	envNamespace = "TENCENTCLOUD_"

	EnvSecretID  = envNamespace + "SECRET_ID"
	EnvSecretKey = envNamespace + "SECRET_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	SecretID  string
	SecretKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

type DomainData struct {
	domain    string
	subDomain string
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
	client *dnspod.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Tencent Cloud DNS.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	values, err := env.Get(EnvSecretID, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("tencentcloud: %w", err)
	}

	config.SecretID = values[EnvSecretID]
	config.SecretKey = values[EnvSecretKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for alidns.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("tencentcloud: the configuration of the DNS provider is nil")
	}

	credential := common.NewCredential(
		config.SecretID,
		config.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	client, _ := dnspod.NewClient(credential, "", cpf)

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
	domainData := getDomainData(fqdn)

	err := d.CreateRecordData(domainData, value)
	if err != nil {
		return fmt.Errorf("tencentcloud: API call failed: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	domainData := getDomainData(fqdn)

	records, err := d.ListRecordData(domainData)
	if err != nil {
		return fmt.Errorf("tencentcloud: API call failed: %w", err)
	}

	for _, item := range records {
		err := d.DeleteRecordData(domainData, item)
		if err != nil {
			return fmt.Errorf("tencentcloud: API call failed: %w", err)
		}
	}

	return nil
}

func getDomainData(str string) *DomainData {
	domainRegexp := regexp.MustCompile(`(.*)\.(.*\..*)\.`)
	params := domainRegexp.FindStringSubmatch(str)
	return &DomainData{domain: params[len(params)-1], subDomain: params[len(params)-2]}
}

func (d *DNSProvider) CreateRecordData(domainData *DomainData, value string) error {
	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(domainData.domain)
	request.SubDomain = common.StringPtr(domainData.subDomain)
	request.RecordType = common.StringPtr("TXT")
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(value)
	request.TTL = common.Uint64Ptr(uint64(d.config.TTL))

	_, err := d.client.CreateRecord(request)
	if err != nil {
		return err
	}
	return nil
}

func (d *DNSProvider) ListRecordData(domainData *DomainData) ([]*dnspod.RecordListItem, error) {
	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(domainData.domain)
	request.Subdomain = common.StringPtr(domainData.subDomain)
	request.RecordType = common.StringPtr("TXT")

	response, err := d.client.DescribeRecordList(request)
	if err != nil {
		return nil, err
	}
	return response.Response.RecordList, nil
}

func (d *DNSProvider) DeleteRecordData(domainData *DomainData, item *dnspod.RecordListItem) error {
	request := dnspod.NewDeleteRecordRequest()
	request.Domain = common.StringPtr(domainData.domain)
	request.RecordId = item.RecordId
	_, err := d.client.DeleteRecord(request)
	if err != nil {
		return err
	}
	return nil
}
