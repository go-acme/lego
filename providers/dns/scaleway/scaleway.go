// Package scaleway implements a DNS provider for solving the DNS-01 challenge using Scaleway Domains API.
// Scaleway Domain API reference: https://developers.scaleway.com/en/products/domain/api/
// Token: https://www.scaleway.com/en/docs/generate-an-api-token/
package scaleway

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	scalewayDNS "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	minTTL                    = 60
	defaultPollingInterval    = 10 * time.Second
	defaultPropagationTimeout = 120 * time.Second
)

// Environment variables names.
const (
	envNamespace = "SCALEWAY_"

	EnvAPIToken  = envNamespace + "API_TOKEN"
	EnvProjectID = envNamespace + "PROJECT_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ProjectID          string
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, defaultPollingInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *scalewayDNS.API
}

// NewDNSProvider returns a DNSProvider instance configured for Scaleway Domains API.
// API token must be passed in the environment variable SCALEWAY_API_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("scaleway: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvAPIToken]
	config.ProjectID = env.GetOrDefaultString(EnvProjectID, "")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for scaleway.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("scaleway: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("scaleway: credentials missing")
	}

	if config.TTL < minTTL {
		config.TTL = minTTL
	}

	configuration := []scw.ClientOption{scw.WithAuth("SCWXXXXXXXXXXXXXXXXX", config.Token), scw.WithUserAgent("Scaleway Lego's provider")}

	if config.ProjectID != "" {
		configuration = append(configuration, scw.WithDefaultProjectID(config.ProjectID))
	}

	// Create a Scaleway client
	clientScw, err := scw.NewClient(configuration...)
	if err != nil {
		return nil, err
	}

	return &DNSProvider{config: config, client: scalewayDNS.NewAPI(clientScw)}, nil
}

// Timeout returns the Timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill DNS-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	comment := "used by lego"
	request := scalewayDNS.RecordChange{
		Add: &scalewayDNS.RecordChangeAdd{
			Records: []*scalewayDNS.Record{
				{
					Data:    fmt.Sprintf(`"%s"`, value),
					Name:    fqdn,
					TTL:     uint32(d.config.TTL),
					Type:    scalewayDNS.RecordTypeTXT,
					Comment: &comment,
				},
			},
		},
	}

	returnAllRecords := false
	_, err := d.client.UpdateDNSZoneRecords(
		&scalewayDNS.UpdateDNSZoneRecordsRequest{
			DNSZone:          domain,
			Changes:          []*scalewayDNS.RecordChange{&request},
			ReturnAllRecords: &returnAllRecords,
		},
	)
	if err != nil {
		return fmt.Errorf("scaleway: %w", err)
	}
	return nil
}

// CleanUp removes a TXT record used for DNS-01 challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	data := fmt.Sprintf(`"%s"`, value)

	request := scalewayDNS.RecordChange{
		Delete: &scalewayDNS.RecordChangeDelete{
			IDFields: &scalewayDNS.RecordIdentifier{
				Name: fqdn,
				Type: scalewayDNS.RecordTypeTXT,
				Data: &data,
			},
		},
	}

	returnAllRecords := false
	_, err := d.client.UpdateDNSZoneRecords(
		&scalewayDNS.UpdateDNSZoneRecordsRequest{
			DNSZone:          domain,
			Changes:          []*scalewayDNS.RecordChange{&request},
			ReturnAllRecords: &returnAllRecords,
		},
	)
	if err != nil {
		return fmt.Errorf("scaleway: %w", err)
	}
	return nil
}
