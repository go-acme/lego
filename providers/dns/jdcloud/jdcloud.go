// Package jdcloud implements a DNS provider for solving the DNS-01 challenge using JD Cloud.
package jdcloud

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-acme/jdcloud-sdk-go/core"
	"github.com/go-acme/jdcloud-sdk-go/services/domainservice/apis"
	jdcclient "github.com/go-acme/jdcloud-sdk-go/services/domainservice/client"
	domainservice "github.com/go-acme/jdcloud-sdk-go/services/domainservice/models"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "JDCLOUD_"

	EnvAccessKeyID     = envNamespace + "ACCESS_KEY_ID"
	EnvAccessKeySecret = envNamespace + "ACCESS_KEY_SECRET"
	EnvRegionID        = envNamespace + "REGION_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	RegionID        string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *jdcclient.DomainserviceClient

	recordIDs   map[string]int
	domainIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for JD Cloud.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessKeyID, EnvAccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("jdcloud: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessKeyID = values[EnvAccessKeyID]
	config.AccessKeySecret = values[EnvAccessKeySecret]

	// https://docs.jdcloud.com/en/common-declaration/api/introduction#Region%20Code
	config.RegionID = env.GetOrDefaultString(EnvRegionID, "cn-north-1")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for JD Cloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("jdcloud: the configuration of the DNS provider is nil")
	}

	if config.AccessKeyID == "" || config.AccessKeySecret == "" {
		return nil, errors.New("jdcloud: missing credentials")
	}

	cred := core.NewCredentials(config.AccessKeyID, config.AccessKeySecret)

	client := jdcclient.NewDomainserviceClient(cred)
	client.DisableLogger()
	client.Config.SetTimeout(config.HTTPTimeout)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int),
		domainIDs: make(map[string]int),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("jdcloud: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("jdcloud: %w", err)
	}

	zone, err := d.findZone(dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("jdcloud: %w", err)
	}

	// https://docs.jdcloud.com/cn/jd-cloud-dns/api/createresourcerecord
	crrr := apis.NewCreateResourceRecordRequestWithAllParams(
		d.config.RegionID,
		strconv.Itoa(zone.Id),
		&domainservice.AddRR{
			HostRecord: subDomain,
			HostValue:  info.Value,
			Ttl:        d.config.TTL,
			Type:       "TXT",
			ViewValue:  -1,
		},
	)

	record, err := jdcclient.CreateResourceRecord(d.client, crrr)
	if err != nil {
		return fmt.Errorf("jdcloud: create resource record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.domainIDs[token] = zone.Id
	d.recordIDs[token] = record.Result.DataList.Id
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, recordOK := d.recordIDs[token]
	domainID, domainOK := d.domainIDs[token]
	d.recordIDsMu.Unlock()

	if !recordOK {
		return fmt.Errorf("jdcloud: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	if !domainOK {
		return fmt.Errorf("jdcloud: unknown domain ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	// https://docs.jdcloud.com/cn/jd-cloud-dns/api/deleteresourcerecord
	drrr := apis.NewDeleteResourceRecordRequestWithAllParams(
		d.config.RegionID,
		strconv.Itoa(domainID),
		strconv.Itoa(recordID),
	)

	_, err := jdcclient.DeleteResourceRecord(d.client, drrr)
	if err != nil {
		return fmt.Errorf("jdcloud: delete resource record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(zone string) (*domainservice.DomainInfo, error) {
	// https://docs.jdcloud.com/cn/jd-cloud-dns/api/describedomains
	ddr := apis.NewDescribeDomainsRequestWithoutParam()
	ddr.SetRegionId(d.config.RegionID)
	ddr.SetPageNumber(1)
	ddr.SetPageSize(10)
	ddr.SetDomainName(zone)

	for {
		response, err := jdcclient.DescribeDomains(d.client, ddr)
		if err != nil {
			return nil, fmt.Errorf("describe domains: %w", err)
		}

		for _, d := range response.Result.DataList {
			if d.DomainName == zone {
				return &d, nil
			}
		}

		if len(response.Result.DataList) < ddr.PageSize || response.Result.TotalPage <= ddr.PageNumber {
			break
		}

		ddr.SetPageNumber(ddr.PageNumber + 1)
	}

	return nil, errors.New("zone not found")
}
