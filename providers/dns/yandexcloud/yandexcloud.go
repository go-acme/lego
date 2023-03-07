// Package yandexcloud implements a DNS provider for solving the DNS-01 challenge using Yandex Cloud.
package yandexcloud

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	ycdns "github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
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
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	IamToken string
	FolderID string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *ycsdk.SDK
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
		return nil, fmt.Errorf("yandexcloud: some credentials information are missing IAM token")
	}

	if config.FolderID == "" {
		return nil, fmt.Errorf("yandexcloud: some credentials information are missing folder id")
	}

	creds, err := decodeCredentials(config.IamToken)
	if err != nil {
		return nil, fmt.Errorf("yandexcloud: iam token is malformed: %w", err)
	}

	client, err := ycsdk.Build(context.Background(), ycsdk.Config{Credentials: creds})
	if err != nil {
		return nil, errors.New("yandexcloud: unable to build yandex cloud sdk")
	}

	return &DNSProvider{
		client: client,
		config: config,
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (r *DNSProvider) Present(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	ctx := context.Background()

	zones, err := r.getZones(ctx)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	var zoneID string

	for _, zone := range zones {
		if zone.GetZone() == authZone {
			zoneID = zone.GetId()
		}
	}

	if zoneID == "" {
		return fmt.Errorf("yandexcloud: cant find dns zone %s in yandex cloud", authZone)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	err = r.upsertRecordSetData(ctx, zoneID, subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (r *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	ctx := context.Background()

	zones, err := r.getZones(ctx)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	var zoneID string

	for _, zone := range zones {
		if zone.GetZone() == authZone {
			zoneID = zone.GetId()
		}
	}

	if zoneID == "" {
		return nil
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	err = r.removeRecordSetData(ctx, zoneID, subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (r *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return r.config.PropagationTimeout, r.config.PollingInterval
}

// getZones retrieves available zones from yandex cloud.
func (r *DNSProvider) getZones(ctx context.Context) ([]*ycdns.DnsZone, error) {
	list := &ycdns.ListDnsZonesRequest{
		FolderId: r.config.FolderID,
	}

	response, err := r.client.DNS().DnsZone().List(ctx, list)
	if err != nil {
		return nil, errors.New("unable to fetch dns zones")
	}

	return response.DnsZones, nil
}

func (r *DNSProvider) upsertRecordSetData(ctx context.Context, zoneID, name, value string) error {
	get := &ycdns.GetDnsZoneRecordSetRequest{
		DnsZoneId: zoneID,
		Name:      name,
		Type:      "TXT",
	}

	exist, err := r.client.DNS().DnsZone().GetRecordSet(ctx, get)
	if err != nil {
		if !strings.Contains(err.Error(), "RecordSet not found") {
			return err
		}
	}

	record := &ycdns.RecordSet{
		Name: name,
		Type: "TXT",
		Ttl:  int64(r.config.TTL),
		Data: []string{},
	}

	var deletions []*ycdns.RecordSet
	if exist != nil {
		record.Data = append(record.Data, exist.Data...)
		deletions = append(deletions, exist)
	}

	appended := appendRecordSetData(record, value)
	if !appended {
		// The value already present in RecordSet, nothing to do
		return nil
	}

	update := &ycdns.UpdateRecordSetsRequest{
		DnsZoneId: zoneID,
		Deletions: deletions,
		Additions: []*ycdns.RecordSet{record},
	}

	_, err = r.client.DNS().DnsZone().UpdateRecordSets(ctx, update)

	return err
}

func (r *DNSProvider) removeRecordSetData(ctx context.Context, zoneID, name, value string) error {
	get := &ycdns.GetDnsZoneRecordSetRequest{
		DnsZoneId: zoneID,
		Name:      name,
		Type:      "TXT",
	}

	previousRecord, err := r.client.DNS().DnsZone().GetRecordSet(ctx, get)
	if err != nil {
		if strings.Contains(err.Error(), "RecordSet not found") {
			// RecordSet is not present, nothing to do
			return nil
		}

		return err
	}

	var additions []*ycdns.RecordSet

	if len(previousRecord.Data) > 1 {
		// RecordSet is not empty we should update it
		record := &ycdns.RecordSet{
			Name: name,
			Type: "TXT",
			Ttl:  int64(r.config.TTL),
			Data: []string{},
		}

		for _, data := range previousRecord.Data {
			if data != value {
				record.Data = append(record.Data, data)
			}
		}

		additions = append(additions, record)
	}

	update := &ycdns.UpdateRecordSetsRequest{
		DnsZoneId: zoneID,
		Deletions: []*ycdns.RecordSet{previousRecord},
		Additions: additions,
	}

	_, err = r.client.DNS().DnsZone().UpdateRecordSets(ctx, update)

	return err
}

// decodeCredentials converts base64 encoded json of iam token to struct.
func decodeCredentials(accountB64 string) (ycsdk.Credentials, error) {
	account, err := base64.StdEncoding.DecodeString(accountB64)
	if err != nil {
		return nil, err
	}

	key := &iamkey.Key{}
	err = json.Unmarshal(account, key)
	if err != nil {
		return nil, err
	}

	return ycsdk.ServiceAccountKey(key)
}

func appendRecordSetData(record *ycdns.RecordSet, value string) bool {
	for _, data := range record.Data {
		if data == value {
			return false
		}
	}

	record.Data = append(record.Data, value)

	return true
}
