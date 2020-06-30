// Package edgedns replaces fastdns, implementing a DNS provider for solving the DNS-01 challenge using Akamai EdgeDNS.
package edgedns

import (
	"errors"
	"fmt"
	"time"

	configdns "github.com/akamai/AkamaiOPEN-edgegrid-golang/configdns-v2"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "AKAMAI_"

	EnvHost         = envNamespace + "HOST"
	EnvClientToken  = envNamespace + "CLIENT_TOKEN"
	EnvClientSecret = envNamespace + "CLIENT_SECRET"
	EnvAccessToken  = envNamespace + "ACCESS_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"

	AkamaiDefaultPropagationTimeout = 3 * time.Minute  // 3 minutes
	AkamaiDefaultPollInterval       = 15 * time.Second // 15 seconds
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	edgegrid.Config
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, AkamaiDefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, AkamaiDefaultPollInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider uses the supplied environment variables to return a DNSProvider instance:
// AKAMAI_HOST, AKAMAI_CLIENT_TOKEN, AKAMAI_CLIENT_SECRET, AKAMAI_ACCESS_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvHost, EnvClientToken, EnvClientSecret, EnvAccessToken)
	if err != nil {
		return nil, fmt.Errorf("edgedns: %w", err)
	}

	config := NewDefaultConfig()
	config.Config = edgegrid.Config{
		Host:         values[EnvHost],
		ClientToken:  values[EnvClientToken],
		ClientSecret: values[EnvClientSecret],
		AccessToken:  values[EnvAccessToken],
		MaxBody:      131072,
	}
	configdns.Init(config.Config)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for EdgeDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("edgedns: the configuration of the DNS provider is nil")
	}

	if config.ClientToken == "" || config.ClientSecret == "" || config.AccessToken == "" || config.Host == "" {
		return nil, errors.New("edgedns: credentials are missing")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	zoneName, recordName, err := d.findZoneAndRecordName(fqdn, domain)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}
	recordName = recordName + "." + zoneName

	existingRec, err := configdns.GetRecord(zoneName, recordName, "TXT")
	if err == nil {
		if existingRec == nil {
			return fmt.Errorf("edgedns: Unknown error")
		}
	} else {
		if configdns.IsConfigDNSError(err) && err.(configdns.ConfigDNSError).NotFound() {
			target := make([]string, 0, 1)
			target = append(target, value)
			record := &configdns.RecordBody{Name: recordName,
				RecordType: "TXT",
				TTL:        d.config.TTL,
				Target:     target}
			return record.Save(zoneName)
		}
		return fmt.Errorf("edgedns: %w", err)
	}

	return nil
}

// CleanUp removes the record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	zoneName, recordName, err := d.findZoneAndRecordName(fqdn, domain)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}
	recordName = recordName + "." + zoneName
	existingRec, err := configdns.GetRecord(zoneName, recordName, "TXT")
	if err == nil {
		if existingRec != nil {
			return existingRec.Delete(zoneName)
		}
		return fmt.Errorf("edgedns: Unknown error")
	}
	if !configdns.IsConfigDNSError(err) || !err.(configdns.ConfigDNSError).NotFound() {
		return fmt.Errorf("edgedns: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZoneAndRecordName(fqdn, domain string) (string, string, error) {
	zone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return "", "", err
	}
	zone = dns01.UnFqdn(zone)
	name := dns01.UnFqdn(fqdn)
	name = name[:len(name)-len("."+zone)]

	return zone, name, nil
}
