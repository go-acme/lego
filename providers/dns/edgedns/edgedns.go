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

	record, err := configdns.GetRecord(zoneName, recordName, "TXT")
	if err != nil && (!configdns.IsConfigDNSError(err) || !err.(configdns.ConfigDNSError).NotFound()) {
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

		return updateRecordset(record, zoneName)
	}

	record = &configdns.RecordBody{
		Name:       recordName,
		RecordType: "TXT",
		TTL:        d.config.TTL,
		Target:     []string{`"` + value + `"`},
	}

	err = record.Save(zoneName)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	return nil
}

// CleanUp removes the record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zoneName, recordName, err := d.findZoneAndRecordName(fqdn, domain)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	recordName = recordName + "." + zoneName

	existingRec, err := configdns.GetRecord(zoneName, recordName, "TXT")
	if err != nil {
		if configdns.IsConfigDNSError(err) && err.(configdns.ConfigDNSError).NotFound() {
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
		return updateRecordset(existingRec, zoneName)
	}

	log.Infof("[DEBUG] Deleting TxtRecord")

	err = existingRec.Delete(zoneName)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	return nil
}

func updateRecordset(rec *configdns.RecordBody, zoneName string) error {
	log.Infof("[DEBUG] Updating TxtRecord: %v", rec.Target)

	err := rec.Update(zoneName)
	if err != nil {
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

func containsValue(tslice []string, value string) bool {
	if len(tslice) == 0 {
		return false
	}

	for _, val := range tslice {
		if val == fmt.Sprintf(`"%s"`, value) {
			return true
		}
	}

	return false
}
