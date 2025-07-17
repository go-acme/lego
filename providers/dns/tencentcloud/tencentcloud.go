// Package tencentcloud implements a DNS provider for solving the DNS-01 challenge using Tencent Cloud DNS.
package tencentcloud

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	dnspod "github.com/go-acme/tencentclouddnspod/v20210323"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

// Environment variables names.
const (
	envNamespace = "TENCENTCLOUD_"

	EnvSecretID     = envNamespace + "SECRET_ID"
	EnvSecretKey    = envNamespace + "SECRET_KEY"
	EnvRegion       = envNamespace + "REGION"
	EnvSessionToken = envNamespace + "SESSION_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	SecretID     string
	SecretKey    string
	Region       string
	SessionToken string

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
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *dnspod.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Tencent Cloud DNS.
// Credentials must be passed in the environment variable: TENCENTCLOUD_SECRET_ID, TENCENTCLOUD_SECRET_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvSecretID, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("tencentcloud: %w", err)
	}

	config := NewDefaultConfig()
	config.SecretID = values[EnvSecretID]
	config.SecretKey = values[EnvSecretKey]
	config.Region = env.GetOrDefaultString(EnvRegion, "")
	config.SessionToken = env.GetOrDefaultString(EnvSessionToken, "")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Tencent Cloud DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("tencentcloud: the configuration of the DNS provider is nil")
	}

	var credential *common.Credential

	switch {
	case config.SecretID != "" && config.SecretKey != "" && config.SessionToken != "":
		credential = common.NewTokenCredential(config.SecretID, config.SecretKey, config.SessionToken)
	case config.SecretID != "" && config.SecretKey != "":
		credential = common.NewCredential(config.SecretID, config.SecretKey)
	default:
		return nil, errors.New("tencentcloud: credentials missing")
	}

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	cpf.HttpProfile.ReqTimeout = int(math.Round(config.HTTPTimeout.Seconds()))

	client, err := dnspod.NewClient(credential, config.Region, cpf)
	if err != nil {
		return nil, fmt.Errorf("tencentcloud: %w", err)
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

	ctx := context.Background()

	zone, err := d.getHostedZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to get hosted zone: %w", err)
	}

	recordName, err := extractRecordName(info.EffectiveFQDN, *zone.Name)
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to extract record name: %w", err)
	}

	request := dnspod.NewCreateRecordRequest()
	request.Domain = zone.Name
	request.DomainId = zone.DomainId
	request.SubDomain = common.StringPtr(recordName)
	request.RecordType = common.StringPtr("TXT")
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(info.Value)
	request.TTL = common.Uint64Ptr(uint64(d.config.TTL))

	_, err = dnspod.CreateRecordWithContext(ctx, d.client, request)
	if err != nil {
		return fmt.Errorf("dnspod: API call failed: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zone, err := d.getHostedZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to get hosted zone: %w", err)
	}

	records, err := d.findTxtRecords(ctx, zone, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to find TXT records: %w", err)
	}

	for _, record := range records {
		request := dnspod.NewDeleteRecordRequest()
		request.Domain = zone.Name
		request.DomainId = zone.DomainId
		request.RecordId = record.RecordId

		_, err := dnspod.DeleteRecordWithContext(ctx, d.client, request)
		if err != nil {
			return fmt.Errorf("tencentcloud: delete record failed: %w", err)
		}
	}

	return nil
}
