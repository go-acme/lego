// Package vkcloud implements a DNS provider for solving the DNS-01 challenge using Yandex Cloud.
package vkcloud

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/gophercloud/gophercloud"
)

const (
	defaultTTL = 60

	defaultIdentityEndpoint = "https://infra.mail.ru/identity/v3/"
	defaultDomainName       = "users"

	defaultDNSEndpoint = "https://mcs.mail.ru/public-dns"
)

// Environment variables names.
const (
	envNamespace = "VK_CLOUD_"

	EnvDNSEndpoint = envNamespace + "DNS_ENDPOINT"

	EnvIdentityEndpoint = envNamespace + "IDENTITY_ENDPOINT"
	EnvDomainName       = envNamespace + "DOMAIN_NAME"

	EnvProjectID = envNamespace + "PROJECT_ID"
	EnvUsername  = envNamespace + "USERNAME"
	EnvPassword  = envNamespace + "PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ProjectID string
	Username  string
	Password  string

	DNSEndpoint string

	IdentityEndpoint string
	DomainName       string

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

type DNSProvider struct {
	client *Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for VK Cloud.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvProjectID, EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("vkcloud: %w", err)
	}

	config := NewDefaultConfig()
	config.ProjectID = values[EnvProjectID]
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]
	config.IdentityEndpoint = env.GetOrDefaultString(EnvIdentityEndpoint, defaultIdentityEndpoint)
	config.DomainName = env.GetOrDefaultString(EnvDomainName, defaultDomainName)
	config.DNSEndpoint = env.GetOrDefaultString(EnvDNSEndpoint, defaultDNSEndpoint)
	if config.DNSEndpoint == "" {
		return nil, fmt.Errorf("vkcloud: dns api endpoint is missing in credentials information")
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Yandex Cloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("vkcloud: the configuration of the DNS provider is nil")
	}

	authOpts, err := getAuthOptions(config)
	if err != nil {
		return nil, fmt.Errorf("vkcloud: credentials are malformed: %w", err)
	}

	if config.DNSEndpoint == "" {
		return nil, fmt.Errorf("vkcloud: dns endpoint is missing in config")
	}

	client, err := NewClient(config.DNSEndpoint, authOpts)
	if err != nil {
		return nil, fmt.Errorf("vkcloud: unable to build vk cloud client: %w", err)
	}

	return &DNSProvider{
		client: client,
		config: config,
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (r *DNSProvider) Present(domain, _, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("vkcloud: %w", err)
	}
	authZone = strings.TrimRight(authZone, ".")

	zones, err := r.client.ListZones()
	if err != nil {
		return fmt.Errorf("vkcloud: unable to fetch dns zones: %w", err)
	}

	var zoneUUID string

	for _, zone := range zones {
		if zone.Zone == authZone {
			zoneUUID = zone.UUID
		}
	}

	if zoneUUID == "" {
		return fmt.Errorf("vkcloud: cant find dns zone %s in vk cloud", authZone)
	}

	name := fqdn[:len(fqdn)-len(authZone)-1]

	err = r.upsertTXTRecord(zoneUUID, name, value)
	if err != nil {
		return fmt.Errorf("vkcloud: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (r *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("vkcloud: %w", err)
	}
	authZone = strings.TrimRight(authZone, ".")

	zones, err := r.client.ListZones()
	if err != nil {
		return fmt.Errorf("vkcloud: unable to fetch dns zones: %w", err)
	}

	var zoneUUID string

	for _, zone := range zones {
		if zone.Zone == authZone {
			zoneUUID = zone.UUID
		}
	}

	if zoneUUID == "" {
		return nil
	}

	name := fqdn[:len(fqdn)-len(authZone)-1]

	err = r.removeTXTRecord(zoneUUID, name, value)
	if err != nil {
		return fmt.Errorf("vkcloud: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (r *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return r.config.PropagationTimeout, r.config.PollingInterval
}

func (r *DNSProvider) upsertTXTRecord(zoneUUID, name, value string) error {
	records, err := r.client.ListTXTRecords(zoneUUID)
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.Name == name && record.Content == value {
			// The DNSRecord is already present, nothing to do
			return nil
		}
	}

	return r.client.CreateTXTRecord(zoneUUID, &DNSTXTRecord{
		Name:    name,
		Content: value,
		TTL:     r.config.TTL,
	})
}

func (r *DNSProvider) removeTXTRecord(zoneUUID, name, value string) error {
	records, err := r.client.ListTXTRecords(zoneUUID)
	if err != nil {
		return err
	}

	name = strings.TrimRight(name, ".")
	for _, record := range records {
		if record.Name == name && record.Content == value {
			return r.client.DeleteTXTRecord(zoneUUID, record.UUID)
		}
	}

	// The DNSRecord is not present, nothing to do
	return nil
}

func getAuthOptions(config *Config) (gophercloud.AuthOptions, error) {
	opts := gophercloud.AuthOptions{}

	if config.ProjectID == "" {
		return opts, fmt.Errorf("vkcloud: project id is missing in credentials information")
	}

	if config.Username == "" {
		return opts, fmt.Errorf("vkcloud: username is missing in credentials information")
	}

	if config.Password == "" {
		return opts, fmt.Errorf("vkcloud: password is missing in credentials information")
	}

	if config.IdentityEndpoint == "" {
		return opts, fmt.Errorf("vkcloud: identity endpoint is missing in config")
	}

	if config.DomainName == "" {
		return opts, fmt.Errorf("vkcloud: domain name is missing in config")
	}

	opts.IdentityEndpoint = config.IdentityEndpoint
	opts.Username = config.Username
	opts.Password = config.Password
	opts.DomainName = config.DomainName
	opts.TenantID = config.ProjectID

	return opts, nil
}
