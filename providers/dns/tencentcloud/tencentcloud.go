// Package tencentcloud implements a DNS provider for solving the DNS-01 challenge using Tencent Cloud DNS.
package tencentcloud

import (
	"errors"
	"fmt"
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
	EnvRegion    = envNamespace + "REGION"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	SecretID  string
	SecretKey string
	Region    string

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

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Tencent Cloud DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("tencentcloud: the configuration of the DNS provider is nil")
	}

	if config.SecretID == "" || config.SecretKey == "" {
		return nil, errors.New("tencentcloud: credentials missing")
	}

	credential := common.NewCredential(config.SecretID, config.SecretKey)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"

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
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	domainData, err := getDomainData(fqdn)
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to get domain data: %w", err)
	}

	err = d.createRecordData(domainData, value)
	if err != nil {
		return fmt.Errorf("tencentcloud: create record failed: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	domainData, err := getDomainData(fqdn)
	if err != nil {
		return fmt.Errorf("tencentcloud: failed to get domain data: %w", err)
	}

	records, err := d.listRecordData(domainData)
	if err != nil {
		return fmt.Errorf("tencentcloud: list records failed: %w", err)
	}

	for _, item := range records {
		err := d.deleteRecordData(domainData, item)
		if err != nil {
			return fmt.Errorf("tencentcloud: delete record failed: %w", err)
		}
	}

	return nil
}
