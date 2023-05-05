// Package edgedns replaces fastdns, implementing a DNS provider for solving the DNS-01 challenge using Akamai EdgeDNS.
package edgedns

import (
	"errors"
	"fmt"
	"strings"
	"time"

	configdns "github.com/akamai/AkamaiOPEN-edgegrid-golang/configdns-v2"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "AKAMAI_"

	EnvEdgeRc        = envNamespace + "EDGERC"
	EnvEdgeRcSection = envNamespace + "EDGERC_SECTION"

	EnvHost         = envNamespace + "HOST"
	EnvClientToken  = envNamespace + "CLIENT_TOKEN"
	EnvClientSecret = envNamespace + "CLIENT_SECRET"
	EnvAccessToken  = envNamespace + "ACCESS_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

const (
	defaultPropagationTimeout = 3 * time.Minute
	defaultPollInterval       = 15 * time.Second
)

const maxBody = 131072

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
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, defaultPollInterval),
		Config:             edgegrid.Config{MaxBody: maxBody},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Akamai EdgeDNS:
// Akamai credentials are automatically detected in the following locations and prioritized in the following order:
//
// 1. Section-specific environment variables `AKAMAI_{SECTION}_HOST`, `AKAMAI_{SECTION}_ACCESS_TOKEN`, `AKAMAI_{SECTION}_CLIENT_TOKEN`, `AKAMAI_{SECTION}_CLIENT_SECRET` where `{SECTION}` is specified using `AKAMAI_EDGERC_SECTION`
// 2. If `AKAMAI_EDGERC_SECTION` is not defined or is set to `default`: Environment variables `AKAMAI_HOST`, `AKAMAI_ACCESS_TOKEN`, `AKAMAI_CLIENT_TOKEN`, `AKAMAI_CLIENT_SECRET`
// 3. .edgerc file located at `AKAMAI_EDGERC` (defaults to `~/.edgerc`, sections can be specified using `AKAMAI_EDGERC_SECTION`)
// 4. Default environment variables: `AKAMAI_HOST`, `AKAMAI_ACCESS_TOKEN`, `AKAMAI_CLIENT_TOKEN`, `AKAMAI_CLIENT_SECRET`
//
// See also: https://developer.akamai.com/api/getting-started
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	rcPath := env.GetOrDefaultString(EnvEdgeRc, "")
	rcSection := env.GetOrDefaultString(EnvEdgeRcSection, "")

	conf, err := edgegrid.Init(rcPath, rcSection)
	if err != nil {
		return nil, fmt.Errorf("edgedns: %w", err)
	}

	conf.MaxBody = maxBody

	config.Config = conf

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for EdgeDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("edgedns: the configuration of the DNS provider is nil")
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
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := getZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	record, err := configdns.GetRecord(zone, info.EffectiveFQDN, "TXT")
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("edgedns: %w", err)
	}

	if err == nil && record == nil {
		return fmt.Errorf("edgedns: unknown error")
	}

	if record != nil {
		log.Infof("TXT record already exists. Updating target")

		if containsValue(record.Target, info.Value) {
			// have a record and have entry already
			return nil
		}

		record.Target = append(record.Target, `"`+info.Value+`"`)
		record.TTL = d.config.TTL

		err = record.Update(zone)
		if err != nil {
			return fmt.Errorf("edgedns: %w", err)
		}

		return nil
	}

	record = &configdns.RecordBody{
		Name:       info.EffectiveFQDN,
		RecordType: "TXT",
		TTL:        d.config.TTL,
		Target:     []string{`"` + info.Value + `"`},
	}

	err = record.Save(zone)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	return nil
}

// CleanUp removes the record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := getZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	existingRec, err := configdns.GetRecord(zone, info.EffectiveFQDN, "TXT")
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

	if !containsValue(existingRec.Target, info.Value) {
		return nil
	}

	var newRData []string
	for _, val := range existingRec.Target {
		val = strings.Trim(val, `"`)
		if val == info.Value {
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

		return nil
	}

	err = existingRec.Delete(zone)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	return nil
}

func getZone(domain string) (string, error) {
	zone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return "", fmt.Errorf("could not find zone for FQDN %q: %w", domain, err)
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

	var e configdns.ConfigDNSError
	return errors.As(err, &e) && e.NotFound()
}
