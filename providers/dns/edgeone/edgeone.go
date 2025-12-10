// Package edgeone implements a DNS provider for solving the DNS-01 challenge using Tencent EdgeOne.
package edgeone

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/ptr"
	teo "github.com/go-acme/tencentedgdeone/v20220901"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"golang.org/x/net/idna"
)

// Environment variables names.
const (
	envNamespace = "EDGEONE_"

	EnvSecretID     = envNamespace + "SECRET_ID"
	EnvSecretKey    = envNamespace + "SECRET_KEY"
	EnvRegion       = envNamespace + "REGION"
	EnvSessionToken = envNamespace + "SESSION_TOKEN"
	EnvZonesMapping = envNamespace + "ZONES_MAPPING"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	SecretID     string
	SecretKey    string
	Region       string
	SessionToken string

	ZonesMapping map[string]string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 20*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 30*time.Second),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *teo.Client

	recordIDs   map[string]*string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Tencent EdgeOne.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvSecretID, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("edgeone: %w", err)
	}

	config := NewDefaultConfig()
	config.SecretID = values[EnvSecretID]
	config.SecretKey = values[EnvSecretKey]
	config.Region = env.GetOrDefaultString(EnvRegion, "")
	config.SessionToken = env.GetOrDefaultString(EnvSessionToken, "")

	mapping := env.GetOrDefaultString(EnvZonesMapping, "")
	if mapping != "" {
		config.ZonesMapping, err = env.ParsePairs(mapping)
		if err != nil {
			return nil, fmt.Errorf("edgeone: zones mapping: %w", err)
		}
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Tencent EdgeOne.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("edgeone: the configuration of the DNS provider is nil")
	}

	var credential *common.Credential

	switch {
	case config.SecretID != "" && config.SecretKey != "" && config.SessionToken != "":
		credential = common.NewTokenCredential(config.SecretID, config.SecretKey, config.SessionToken)
	case config.SecretID != "" && config.SecretKey != "":
		credential = common.NewCredential(config.SecretID, config.SecretKey)
	default:
		return nil, errors.New("edgeone: credentials missing")
	}

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "teo.intl.tencentcloudapi.com"
	cpf.HttpProfile.ReqTimeout = int(math.Round(config.HTTPTimeout.Seconds()))

	client, err := teo.NewClient(credential, config.Region, cpf)
	if err != nil {
		return nil, fmt.Errorf("edgeone: %w", err)
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: map[string]*string{},
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zoneID, err := d.getHostedZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("edgeone: failed to get hosted zone: %w", err)
	}

	punnyCoded, err := idna.ToASCII(dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("edgeone: fail to convert punycode: %w", err)
	}

	request := teo.NewCreateDnsRecordRequest()
	request.Name = ptr.Pointer(punnyCoded)
	request.ZoneId = zoneID
	request.Type = ptr.Pointer("TXT")
	request.Content = ptr.Pointer(info.Value)
	request.TTL = ptr.Pointer(int64(d.config.TTL))

	nr, err := teo.CreateDnsRecordWithContext(ctx, d.client, request)
	if err != nil {
		return fmt.Errorf("edgeone: API call failed: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = nr.Response.RecordId
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zoneID, err := d.getHostedZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("edgeone: failed to get hosted zone: %w", err)
	}

	// get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("edgeone: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	request := teo.NewDeleteDnsRecordsRequest()
	request.ZoneId = zoneID
	request.RecordIds = []*string{recordID}

	_, err = teo.DeleteDnsRecordsWithContext(ctx, d.client, request)
	if err != nil {
		return fmt.Errorf("edgeone: delete record failed: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
