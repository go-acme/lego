// Package huaweicloud implements a DNS provider for solving the DNS-01 challenge using Huawei Cloud.
package huaweicloud

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	hwauthbasic "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	hwconfig "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	hwdns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	hwmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	hwregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
)

// Environment variables names.
const (
	envNamespace = "HUAWEICLOUD_"

	EnvAccessKeyID     = envNamespace + "ACCESS_KEY_ID"
	EnvSecretAccessKey = envNamespace + "SECRET_ACCESS_KEY"
	EnvRegion          = envNamespace + "REGION"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int32
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                int32(env.GetOrDefaultInt(EnvTTL, 300)),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *hwdns.DnsClient
}

// NewDNSProvider returns a DNSProvider instance configured for Huawei Cloud.
// Credentials must be passed in the environment variables:
// HUAWEICLOUD_ACCESS_KEY_ID, HUAWEICLOUD_SECRET_ACCESS_KEY, and HUAWEICLOUD_REGION.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessKeyID, EnvSecretAccessKey, EnvRegion)
	if err != nil {
		return nil, fmt.Errorf("huaweicloud: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessKeyID = values[EnvAccessKeyID]
	config.SecretAccessKey = values[EnvSecretAccessKey]
	config.Region = values[EnvRegion]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Huawei Cloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("huaweicloud: the configuration of the DNS provider is nil")
	}

	if config.AccessKeyID == "" || config.SecretAccessKey == "" || config.Region == "" {
		return nil, errors.New("huaweicloud: credentials missing")
	}

	auth, err := hwauthbasic.NewCredentialsBuilder().
		WithAk(config.AccessKeyID).
		WithSk(config.SecretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, fmt.Errorf("huaweicloud: crendential build: %w", err)
	}

	region, err := hwregion.SafeValueOf(config.Region)
	if err != nil {
		return nil, fmt.Errorf("huaweicloud: safe region: %w", err)
	}

	if region == nil {
		return nil, fmt.Errorf("huaweicloud: region not found: %s", config.Region)
	}

	client, err := hwdns.DnsClientBuilder().
		WithHttpConfig(hwconfig.DefaultHttpConfig().WithTimeout(config.HTTPTimeout)).
		WithRegion(region).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, fmt.Errorf("huaweicloud: client build: %w", err)
	}

	return &DNSProvider{config: config, client: hwdns.NewDnsClient(client)}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("huaweicloud: could not find zone for domain %q: %w", domain, err)
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return fmt.Errorf("huaweicloud: %w", err)
	}

	request := &hwmodel.CreateRecordSetRequest{
		ZoneId: zoneID,
		Body: &hwmodel.CreateRecordSetRequestBody{
			Name:        info.EffectiveFQDN,
			Description: pointer("Added TXT record for ACME dns-01 challenge using lego client"),
			Type:        "TXT",
			Ttl:         pointer(d.config.TTL),
			Records:     []string{strconv.Quote(info.Value)},
		},
	}

	_, err = d.client.CreateRecordSet(request)
	if err != nil {
		return fmt.Errorf("huaweicloud: create record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("huaweicloud: could not find zone for domain %q: %w", domain, err)
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return fmt.Errorf("huaweicloud: %w", err)
	}

	records, err := d.client.ListRecordSetsByZone(&hwmodel.ListRecordSetsByZoneRequest{
		ZoneId: zoneID,
		Name:   pointer(info.EffectiveFQDN),
	})
	if err != nil {
		return fmt.Errorf("huaweicloud: record list: unable to get record %s for zone %s: %w", info.EffectiveFQDN, domain, err)
	}

	recordSets := deref(records.Recordsets)

	var recordID string
	for _, record := range recordSets {
		if deref(record.Type) == "TXT" &&
			deref(record.Name) == info.EffectiveFQDN &&
			slices.Contains(deref(record.Records), strconv.Quote(info.Value)) {
			recordID = deref(recordSets[0].Id)
		}
	}

	if recordID == "" {
		return errors.New("huaweicloud: missing record ID")
	}

	request := &hwmodel.DeleteRecordSetRequest{
		ZoneId:      zoneID,
		RecordsetId: recordID,
	}

	_, err = d.client.DeleteRecordSet(request)
	if err != nil {
		return fmt.Errorf("huaweicloud: delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getZoneID(authZone string) (string, error) {
	zones, err := d.client.ListPublicZones(&hwmodel.ListPublicZonesRequest{})
	if err != nil {
		return "", fmt.Errorf("unable to get zone: %w", err)
	}

	for _, zone := range deref(zones.Zones) {
		if deref(zone.Name) == authZone {
			return deref(zone.Id), nil
		}
	}

	return "", fmt.Errorf("zone %q not found", authZone)
}

func pointer[T string | int | int32 | int64](v T) *T { return &v }

func deref[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}

	return *v
}
