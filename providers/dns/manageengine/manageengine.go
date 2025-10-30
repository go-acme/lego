// Package manageengine implements a DNS provider for solving the DNS-01 challenge using ManageEngine CloudDNS.
package manageengine

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/manageengine/internal"
)

// Environment variables names.
const (
	envNamespace = "MANAGEENGINE_"

	EnvClientID     = envNamespace + "CLIENT_ID"
	EnvClientSecret = envNamespace + "CLIENT_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ClientID     string
	ClientSecret string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for ManageEngine CloudDNS.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvClientID, EnvClientSecret)
	if err != nil {
		return nil, fmt.Errorf("manageengine: %w", err)
	}

	config := NewDefaultConfig()
	config.ClientID = values[EnvClientID]
	config.ClientSecret = values[EnvClientSecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ManageEngine CloudDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("manageengine: the configuration of the DNS provider is nil")
	}

	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, errors.New("manageengine: credentials missing")
	}

	return &DNSProvider{
		config: config,
		client: internal.NewClient(
			clientdebug.Wrap(
				internal.CreateOAuthClient(context.Background(), config.ClientID, config.ClientSecret),
			),
		),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("manageengine: could not find zone for domain %q: %w", domain, err)
	}

	zoneID, err := d.findZoneID(ctx, authZone)
	if err != nil {
		return fmt.Errorf("manageengine: find zone ID: %w", err)
	}

	zoneRecord, err := d.findZoneRecord(ctx, zoneID, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("manageengine: find zone record: %w", err)
	}

	// Update the existing zone record.
	if zoneRecord != nil {
		for _, record := range zoneRecord.Records {
			if slices.Contains(record.Values, info.Value) {
				continue
			}

			zr := internal.ZoneRecord{
				ZoneID:         zoneID,
				SpfTxtDomainID: zoneRecord.SpfTxtDomainID,
				DomainName:     info.EffectiveFQDN,
				DomainTTL:      d.config.TTL,
				RecordType:     "TXT",
				Records: []internal.Record{{
					Values:   append(record.Values, info.Value),
					DomainID: zoneRecord.SpfTxtDomainID,
				}},
			}

			// Update the zone record.
			err = d.client.UpdateZoneRecord(ctx, zr)
			if err != nil {
				return fmt.Errorf("manageengine: update zone record: %w", err)
			}

			return nil
		}

		return errors.New("manageengine: zone already contains the TXT record value")
	}

	// Create a new zone record.
	record := internal.ZoneRecord{
		ZoneID:     zoneID,
		DomainName: info.EffectiveFQDN,
		DomainTTL:  d.config.TTL,
		RecordType: "TXT",
		Records: []internal.Record{{
			Values: []string{info.Value},
		}},
	}

	err = d.client.CreateZoneRecord(ctx, zoneID, record)
	if err != nil {
		return fmt.Errorf("manageengine: create zone record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("manageengine: could not find zone for domain %q: %w", domain, err)
	}

	zoneID, err := d.findZoneID(ctx, authZone)
	if err != nil {
		return fmt.Errorf("manageengine: find zone ID: %w", err)
	}

	zoneRecord, err := d.findZoneRecord(ctx, zoneID, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("manageengine: find zone record: %w", err)
	}

	for _, record := range zoneRecord.Records {
		if !slices.Contains(record.Values, info.Value) {
			continue
		}

		// Delete the zone record.
		if len(record.Values) <= 1 {
			err = d.client.DeleteZoneRecord(ctx, zoneID, zoneRecord.SpfTxtDomainID)
			if err != nil {
				return fmt.Errorf("manageengine: delete zone record: %w", err)
			}

			return nil
		}

		// Update the zone record.
		var values []string

		for _, value := range record.Values {
			if value != info.Value {
				values = append(values, value)
			}
		}

		zr := internal.ZoneRecord{
			ZoneID:         zoneID,
			SpfTxtDomainID: zoneRecord.SpfTxtDomainID,
			DomainName:     info.EffectiveFQDN,
			DomainTTL:      d.config.TTL,
			RecordType:     "TXT",
			Records: []internal.Record{{
				Values:   values,
				DomainID: zoneRecord.SpfTxtDomainID,
			}},
		}

		err = d.client.UpdateZoneRecord(ctx, zr)
		if err != nil {
			return fmt.Errorf("manageengine: create zone record: %w", err)
		}

		return nil
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZoneID(ctx context.Context, authZone string) (int, error) {
	zones, err := d.client.GetAllZones(ctx)
	if err != nil {
		return 0, fmt.Errorf("get all zone groups: %w", err)
	}

	for _, zone := range zones {
		if strings.EqualFold(zone.ZoneName, authZone) {
			return zone.ZoneID, nil
		}
	}

	return 0, fmt.Errorf("zone not found %s", authZone)
}

func (d *DNSProvider) findZoneRecord(ctx context.Context, zoneID int, fqdn string) (*internal.ZoneRecord, error) {
	zoneRecords, err := d.client.GetAllZoneRecords(ctx, zoneID)
	if err != nil {
		return nil, fmt.Errorf("get all zone records: %w", err)
	}

	for _, zoneRecord := range zoneRecords {
		if !strings.EqualFold(zoneRecord.DomainName, fqdn) {
			continue
		}

		if strings.EqualFold(zoneRecord.RecordType, "TXT") {
			return &zoneRecord, nil
		}
	}

	return nil, nil
}
