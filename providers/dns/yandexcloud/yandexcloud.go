// Package yandexcloud implements a DNS provider for solving the DNS-01 challenge using Yandex Cloud.
package yandexcloud

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	ycdnsproto "github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	ycdns "github.com/yandex-cloud/go-sdk/services/dns/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/iamkey"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options"
)

// Environment variables names.
const (
	envNamespace = "YANDEX_CLOUD_"

	EnvIamToken = envNamespace + "IAM_TOKEN"
	EnvFolderID = envNamespace + "FOLDER_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

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
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client ycdns.DnsZoneClient
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
		return nil, errors.New("yandexcloud: some credentials information are missing IAM token")
	}

	if config.FolderID == "" {
		return nil, errors.New("yandexcloud: some credentials information are missing folder id")
	}

	creds, err := decodeCredentials(config.IamToken)
	if err != nil {
		return nil, fmt.Errorf("yandexcloud: iam token is malformed: %w", err)
	}

	sdk, err := ycsdk.Build(context.Background(), options.WithCredentials(creds))
	if err != nil {
		return nil, errors.New("yandexcloud: unable to build yandex cloud sdk")
	}

	return &DNSProvider{
		client: ycdns.NewDnsZoneClient(sdk),
		config: config,
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("yandexcloud: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	zones, err := d.getZones(ctx)
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

	err = d.upsertRecordSetData(ctx, zoneID, subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("yandexcloud: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	zones, err := d.getZones(ctx)
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

	err = d.removeRecordSetData(ctx, zoneID, subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("yandexcloud: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// getZones retrieves available zones from yandex cloud.
func (d *DNSProvider) getZones(ctx context.Context) ([]*ycdnsproto.DnsZone, error) {
	list := &ycdnsproto.ListDnsZonesRequest{
		FolderId: d.config.FolderID,
	}

	response, err := d.client.List(ctx, list)
	if err != nil {
		return nil, errors.New("unable to fetch dns zones")
	}

	return response.GetDnsZones(), nil
}

func (d *DNSProvider) upsertRecordSetData(ctx context.Context, zoneID, name, value string) error {
	get := &ycdnsproto.GetDnsZoneRecordSetRequest{
		DnsZoneId: zoneID,
		Name:      name,
		Type:      "TXT",
	}

	exist, err := d.client.GetRecordSet(ctx, get)
	if err != nil {
		if !strings.Contains(err.Error(), "RecordSet not found") {
			return err
		}
	}

	record := &ycdnsproto.RecordSet{
		Name: name,
		Type: "TXT",
		Ttl:  int64(d.config.TTL),
		Data: []string{},
	}

	var deletions []*ycdnsproto.RecordSet
	if exist != nil {
		record.SetData(append(record.GetData(), exist.GetData()...))
		deletions = append(deletions, exist)
	}

	appended := appendRecordSetData(record, value)
	if !appended {
		// The value already present in RecordSet, nothing to do
		return nil
	}

	update := &ycdnsproto.UpdateRecordSetsRequest{
		DnsZoneId: zoneID,
		Deletions: deletions,
		Additions: []*ycdnsproto.RecordSet{record},
	}

	_, err = d.client.UpdateRecordSets(ctx, update)

	return err
}

func (d *DNSProvider) removeRecordSetData(ctx context.Context, zoneID, name, value string) error {
	get := &ycdnsproto.GetDnsZoneRecordSetRequest{
		DnsZoneId: zoneID,
		Name:      name,
		Type:      "TXT",
	}

	previousRecord, err := d.client.GetRecordSet(ctx, get)
	if err != nil {
		if strings.Contains(err.Error(), "RecordSet not found") {
			// RecordSet is not present, nothing to do
			return nil
		}

		return err
	}

	var additions []*ycdnsproto.RecordSet

	if len(previousRecord.GetData()) > 1 {
		// RecordSet is not empty we should update it
		record := &ycdnsproto.RecordSet{
			Name: name,
			Type: "TXT",
			Ttl:  int64(d.config.TTL),
			Data: []string{},
		}

		for _, data := range previousRecord.GetData() {
			if data != value {
				record.SetData(append(record.GetData(), data))
			}
		}

		additions = append(additions, record)
	}

	update := &ycdnsproto.UpdateRecordSetsRequest{
		DnsZoneId: zoneID,
		Deletions: []*ycdnsproto.RecordSet{previousRecord},
		Additions: additions,
	}

	_, err = d.client.UpdateRecordSets(ctx, update)

	return err
}

// decodeCredentials converts base64 encoded json of iam token to struct.
func decodeCredentials(accountB64 string) (credentials.Credentials, error) {
	account, err := base64.StdEncoding.DecodeString(accountB64)
	if err != nil {
		return nil, err
	}

	key := &iamkey.Key{}
	err = json.Unmarshal(account, key)
	if err != nil {
		return nil, err
	}

	return credentials.ServiceAccountKey(key)
}

func appendRecordSetData(record *ycdnsproto.RecordSet, value string) bool {
	if slices.Contains(record.GetData(), value) {
		return false
	}

	record.SetData(append(record.GetData(), value))

	return true
}
