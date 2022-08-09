// Package yandexcloud implements a DNS provider for solving the DNS-01 challenge using Yandex Cloud.
package yandexcloud

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
)

const defaultTTL = 60

// Environment variables names.
const (
	envNamespace = "YANDEX_CLOUD_"

	EnvIamToken = envNamespace + "IAM_TOKEN"
	EnvFolderID = envNamespace + "FOLDER_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	IamToken string
	FolderID string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, dns01.DefaultPropagationTimeout),
	}
}

type DNSProvider struct {
	sdk    *ycsdk.SDK
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Yandex Cloud.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvIamToken, EnvFolderID)
	if err != nil {
		return nil, fmt.Errorf("yandexcloud: %w", err)
	}

	config := NewDefaultConfig()
	config.IamToken = values[EnvIamToken]
	config.FolderID = values[EnvFolderID]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Yandex Cloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("yandexcloud: the configuration of the DNS provider is nil")
	}

	if config.IamToken == "" {
		return nil, fmt.Errorf("yandexcloud: some credentials information are missing iam token")
	}

	if config.FolderID == "" {
		return nil, fmt.Errorf("yandexcloud: some credentials information are missing folder id")
	}

	ycCreds, err := decodeYcCredentials(config.IamToken)
	if err != nil {
		return nil, fmt.Errorf("yandexcloud: iam token is malformed: %w", err)
	}

	sdk, err := ycsdk.Build(context.TODO(), ycsdk.Config{
		Credentials: ycCreds,
	})
	if err != nil {
		return nil, errors.New("yandexcloud: unable to build yandex cloud sdk")
	}

	return &DNSProvider{
		sdk:    sdk,
		config: config,
	}, nil
}

// GetZones retrieves available zones from yandex cloud.
func (r *DNSProvider) GetZones() ([]*dns.DnsZone, error) {
	request := &dns.ListDnsZonesRequest{
		FolderId: r.config.FolderID,
	}

	response, err := r.sdk.DNS().DnsZone().List(context.TODO(), request)
	if err != nil {
		return nil, errors.New("yandexcloud: unable to fetch dns zones")
	}

	return response.DnsZones, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (r *DNSProvider) Present(domain, _, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	ycZones, err := r.GetZones()
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	var ycZoneID string

	for _, zone := range ycZones {
		if zone.GetZone() == authZone {
			ycZoneID = zone.GetId()
		}
	}

	if ycZoneID == "" {
		return fmt.Errorf("yandexcloud: cant find dns zone %s in yandex cloud", authZone)
	}

	name := fqdn[:len(fqdn)-len(authZone)-1]

	err = r.createOrUpdateRecord(ycZoneID, name, value)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	return err
}

// CleanUp removes the TXT record matching the specified parameters.
func (r *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	ycZones, err := r.GetZones()
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	var ycZoneID string

	for _, zone := range ycZones {
		if zone.GetZone() == authZone {
			ycZoneID = zone.GetId()
		}
	}

	if ycZoneID == "" {
		return nil
	}

	name := fqdn[:len(fqdn)-len(authZone)-1]

	_, err = r.removeRecord(ycZoneID, name)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	return nil
}

func (r *DNSProvider) createOrUpdateRecord(zoneID string, name string, value string) error {
	get := &dns.GetDnsZoneRecordSetRequest{
		DnsZoneId: zoneID,
		Name:      name,
		Type:      "TXT",
	}

	exists, _ := r.sdk.DNS().DnsZone().GetRecordSet(context.TODO(), get)

	var deletions []*dns.RecordSet
	if exists != nil {
		deletions = append(deletions, exists)
	}

	update := &dns.UpdateRecordSetsRequest{
		DnsZoneId: zoneID,
		Deletions: deletions,
		Additions: []*dns.RecordSet{
			{
				Name: name,
				Type: "TXT",
				Ttl:  int64(r.config.TTL),
				Data: []string{
					value,
				},
			},
		},
	}

	_, err := r.sdk.DNS().DnsZone().UpdateRecordSets(context.TODO(), update)

	return err
}

func (r *DNSProvider) removeRecord(zoneID string, name string) (*operation.Operation, error) {
	get := &dns.GetDnsZoneRecordSetRequest{
		DnsZoneId: zoneID,
		Name:      name,
		Type:      "TXT",
	}

	exists, _ := r.sdk.DNS().DnsZone().GetRecordSet(context.TODO(), get)

	var deletions []*dns.RecordSet
	if exists != nil {
		deletions = append(deletions, exists)
	}

	update := &dns.UpdateRecordSetsRequest{
		DnsZoneId: zoneID,
		Deletions: deletions,
		Additions: []*dns.RecordSet{},
	}

	return r.sdk.DNS().DnsZone().UpdateRecordSets(context.TODO(), update)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (r *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return r.config.PropagationTimeout, r.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (r *DNSProvider) Sequential() time.Duration {
	return r.config.SequenceInterval
}

// decodeYcCredentials converts base64 encoded json of iam token to struct.
func decodeYcCredentials(ycAccountB64 string) (ycsdk.Credentials, error) {
	ycAccountJSON, err := base64.StdEncoding.DecodeString(ycAccountB64)
	if err != nil {
		return nil, err
	}

	key := &iamkey.Key{}

	err = json.Unmarshal(ycAccountJSON, key)
	if err != nil {
		return nil, err
	}

	return ycsdk.ServiceAccountKey(key)
}
