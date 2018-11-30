package online

import (
	"errors"
	"fmt"
	"github.com/timewasted/linode/dns"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
	"strings"
	"time"
)

const (
	minTTL = 300
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	APIKey          string
	PollingInterval time.Duration
	TTL             int
}

// DNSProvider implements the acme.ChallengeProvider interface.
type DNSProvider struct {
	config *Config
	client *dns.DNS
}

type hostedZoneInfo struct {
	domainID     int
	resourceName string
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PollingInterval: env.GetOrDefaultSecond("ONLINE_POLLING_INTERVAL", 15*time.Second),
		TTL:             env.GetOrDefaultInt("ONLINE_TTL", minTTL),
	}
}

// NewDNSProvider returns a DNSProvider instance configured for Online.
// Credentials must be passed in the environment variable: ONLINE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("ONLINE_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("online: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["ONLINE_API_KEY"]

	return NewDNSProviderConfig(config)
}


// NewDNSProviderConfig return a DNSProvider instance configured for Online.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("online: the configuration of the DNS provider is nil")
	}

	if len(config.APIKey) == 0 {
		return nil, errors.New("online: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("online: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	return &DNSProvider{
		config: config,
		client: dns.New(config.APIKey),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zone, err := d.getHostedZoneInfo(fqdn)
	if err != nil {
		return err
	}

	if _, err = d.client.CreateDomainResourceTXT(zone.domainID, acme.UnFqdn(fqdn), value, d.config.TTL); err != nil {
		return err
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zone, err := d.getHostedZoneInfo(fqdn)
	if err != nil {
		return err
	}

	// Get all TXT records for the specified domain.
	resources, err := d.client.GetResourcesByType(zone.domainID, "TXT")
	if err != nil {
		return err
	}

	// Remove the specified resource, if it exists.
	for _, resource := range resources {
		if resource.Name == zone.resourceName && resource.Target == value {
			resp, err := d.client.DeleteDomainResource(resource.DomainID, resource.ResourceID)
			if err != nil {
				return err
			}

			if resp.ResourceID != resource.ResourceID {
				return errors.New("error deleting resource: resource IDs do not match")
			}
			break
		}
	}

	return nil
}

func (d *DNSProvider) getHostedZoneInfo(fqdn string) (*hostedZoneInfo, error) {
	// Lookup the zone that handles the specified FQDN.
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return nil, err
	}

	resourceName := strings.TrimSuffix(fqdn, "."+authZone)

	// Query the authority zone.
	domain, err := d.client.GetDomain(acme.UnFqdn(authZone))
	if err != nil {
		return nil, err
	}

	return &hostedZoneInfo{
		domainID:     domain.DomainID,
		resourceName: resourceName,
	}, nil
}
