package nicru

import (
	"errors"
	"fmt"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/nicru/internal"
	"net/http"
	"strconv"
	"time"
)

const (
	envNamespace = "NIC_RU_"

	EnvUsername    = envNamespace + "USER"
	EnvPassword    = envNamespace + "PASSWORD"
	EnvServiceId   = envNamespace + "SERVICE_ID"
	EnvSecret      = envNamespace + "SECRET"
	EnvServiceName = envNamespace + "SERVICE_NAME"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"

	defaultTTL                = 30
	defaultPropagationTimeout = 10 * 60 * time.Second
	defaultPollingInterval    = 60 * time.Second
	defaultHttpTimeout        = 30 * time.Second
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	TTL                int
	Username           string
	Password           string
	ServiceId          string
	Secret             string
	Domain             string
	ServiceName        string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, defaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, defaultHttpTimeout),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *internal.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for NIC RU
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword, EnvServiceId, EnvSecret, EnvServiceName)
	if err != nil {
		return nil, fmt.Errorf("nicru: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]
	config.ServiceId = values[EnvServiceId]
	config.Secret = values[EnvSecret]
	config.ServiceName = values[EnvServiceName]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for NIC RU.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("nicru: the configuration of the DNS provider is nil")
	}

	provider := internal.Provider{
		OAuth2ClientID: config.ServiceId,
		OAuth2SecretID: config.Secret,
		Username:       config.Username,
		Password:       config.Password,
		ServiceName:    config.ServiceName,
	}
	client, err := internal.NewClient(&provider)
	if err != nil {
		return nil, fmt.Errorf("nicru: unable to build RU CENTER client: %w", err)
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
		return fmt.Errorf("nicru: %w", err)
	}

	authZone = dns01.UnFqdn(authZone)

	zones, err := r.client.GetZones()
	var zoneUUID string
	for _, zone := range zones {
		if zone.Name == authZone {
			zoneUUID = zone.ID
		}
	}

	if zoneUUID == "" {
		return fmt.Errorf("nicru: cant find dns zone %s in nic.ru", authZone)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("nicru: %w", err)
	}

	err = r.upsertTxtRecord(authZone, subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("nicru: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (r *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("nicru: %w", err)
	}

	authZone = dns01.UnFqdn(authZone)

	zones, err := r.client.GetZones()
	if err != nil {
		return fmt.Errorf("nicru: unable to fetch dns zones: %w", err)
	}

	var zoneUUID string

	for _, zone := range zones {
		if zone.Name == authZone {
			zoneUUID = zone.ID
		}
	}

	if zoneUUID == "" {
		return nil
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("nicru: %w", err)
	}

	err = r.removeTxtRecord(authZone, subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("nicru: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (r *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return r.config.PropagationTimeout, r.config.PollingInterval
}

func (r *DNSProvider) upsertTxtRecord(zone, name, value string) error {
	records, err := r.client.GetTXTRecords(zone)
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.Text == name && record.String == value {
			return nil
		}
	}

	_, err = r.client.AddTxtRecord(zone, name, value, r.config.TTL)
	if err != nil {
		return err
	}
	_, err = r.client.CommitZone(zone)
	return err
}

func (r *DNSProvider) removeTxtRecord(zone, name, value string) error {
	records, err := r.client.GetRecords(zone)
	if err != nil {
		return err
	}

	name = dns01.UnFqdn(name)
	for _, record := range records {
		if record.Txt != nil {
			if record.Name == name && record.Txt.String == value {
				_id, err := strconv.Atoi(record.ID)
				if err != nil {
					return err
				}
				_, err = r.client.DeleteRecord(zone, _id)
				if err != nil {
					return err
				}
			}
		}
	}

	_, err = r.client.CommitZone(zone)
	return err
}
