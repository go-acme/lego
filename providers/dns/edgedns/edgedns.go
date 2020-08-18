// Package edgedns replaces fastdns, implementing a DNS provider for solving the DNS-01 challenge using Akamai EdgeDNS.
package edgedns

import (
	"errors"
	"fmt"
	"strings"
	"time"

	configdns "github.com/akamai/AkamaiOPEN-edgegrid-golang/configdns-v2"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/log"
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

	DefaultPropagationTimeout = 3 * time.Minute
	DefaultPollInterval       = 15 * time.Second
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	edgegrid.Config
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, DefaultPollInterval),
		Config: edgegrid.Config{
			MaxBody: 131072,
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider uses the supplied environment variables to return a DNSProvider instance:
// AKAMAI_HOST, AKAMAI_CLIENT_TOKEN, AKAMAI_CLIENT_SECRET, AKAMAI_ACCESS_TOKEN.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvHost, EnvClientToken, EnvClientSecret, EnvAccessToken)
	if err != nil {
		return nil, fmt.Errorf("edgedns: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.Config.Host = values[EnvHost]
	config.Config.ClientToken = values[EnvClientToken]
	config.Config.ClientSecret = values[EnvClientSecret]
	config.Config.AccessToken = values[EnvAccessToken]

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

	configdns.Init(config.Config)

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := findZone(domain)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	record, err := configdns.GetRecord(zone, fqdn, "TXT")
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("edgedns: %w", err)
	}

	if err == nil && record == nil {
		return fmt.Errorf("edgedns: unknown error")
	}

	if record != nil {
		log.Infof("TXT record already exists. Updating target")

		if containsValue(record.Target, value) {
			// have a record and have entry already
			return nil
		}

		record.Target = append(record.Target, `"`+value+`"`)
		record.TTL = d.config.TTL

		err = record.Update(zone)
		if err != nil {
			return fmt.Errorf("edgedns: %w", err)
		}
	}

	record = &configdns.RecordBody{
		Name:       fqdn,
		RecordType: "TXT",
		TTL:        d.config.TTL,
		Target:     []string{`"` + value + `"`},
	}

	err = record.Save(zone)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	return nil
}

// CleanUp removes the record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := findZone(domain)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	existingRec, err := configdns.GetRecord(zone, fqdn, "TXT")
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return fmt.Errorf("edgedns: %w", err)
	}

	if existingRec == nil {
		return fmt.Errorf("edgedns: unknown failure")
	}

	if len(existingRec.Target) == 0 {
		return fmt.Errorf("edgedns: TXT record is invalid")
	}

	if !containsValue(existingRec.Target, value) {
		return nil
	}

	var newRData []string
	for _, val := range existingRec.Target {
		val = strings.Trim(val, `"`)
		if val == value {
			continue
		}
		newRData = append(newRData, val)
	}

	if len(newRData) > 0 {
		existingRec.Target = newRData

		err = existingRec.Update(zone)
		if err != nil {
			return fmt.Errorf("edgedns: %w", err)
		}
	}

	err = existingRec.Delete(zone)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	return nil
}

func findZone(domain string) (string, error) {
	zone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return "", err
	}

	return dns01.UnFqdn(zone), nil
}

func containsValue(values []string, value string) bool {
	for _, val := range values {
		if strings.Trim(val, `"`) == value {
			return true
		}
	}

	return false
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	e, ok := err.(configdns.ConfigDNSError)
	return ok && e.NotFound()
}
