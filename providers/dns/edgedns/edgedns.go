// Package edgedns replaces fastdns, implementing a DNS provider for solving the DNS-01 challenge using Akamai EdgeDNS.
package edgedns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	edgegriddns "github.com/akamai/AkamaiOPEN-edgegrid-golang/v11/pkg/dns"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v11/pkg/edgegrid"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v11/pkg/session"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "AKAMAI_"

	EnvEdgeRc           = envNamespace + "EDGERC"
	EnvEdgeRcSection    = envNamespace + "EDGERC_SECTION"
	EnvAccountSwitchKey = envNamespace + "ACCOUNT_SWITCH_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Test Environment variables names (unused).
// TODO(ldez): must be moved into test files.
const (
	EnvHost         = envNamespace + "HOST"
	EnvClientToken  = envNamespace + "CLIENT_TOKEN"
	EnvClientSecret = envNamespace + "CLIENT_SECRET"
	EnvAccessToken  = envNamespace + "ACCESS_TOKEN"
)

const (
	defaultPropagationTimeout = 3 * time.Minute
	defaultPollInterval       = 15 * time.Second
)

const maxBody = 131072

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	*edgegrid.Config

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
		Config:             &edgegrid.Config{MaxBody: maxBody},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Akamai EdgeDNS:
// Akamai's credentials are automatically detected in the following locations and prioritized in the following order:
//
// 1. Section-specific environment variables `AKAMAI_{SECTION}_HOST`, `AKAMAI_{SECTION}_ACCESS_TOKEN`, `AKAMAI_{SECTION}_CLIENT_TOKEN`, `AKAMAI_{SECTION}_CLIENT_SECRET` where `{SECTION}` is specified using `AKAMAI_EDGERC_SECTION`
// 2. If `AKAMAI_EDGERC_SECTION` is not defined or is set to `default`: Environment variables `AKAMAI_HOST`, `AKAMAI_ACCESS_TOKEN`, `AKAMAI_CLIENT_TOKEN`, `AKAMAI_CLIENT_SECRET`
// 3. .edgerc file located at `AKAMAI_EDGERC` (defaults to `~/.edgerc`, sections can be specified using `AKAMAI_EDGERC_SECTION`)
//
// See also: https://developer.akamai.com/api/getting-started
func NewDNSProvider() (*DNSProvider, error) {
	conf, err := edgegrid.New(
		edgegrid.WithEnv(true),
		edgegrid.WithFile(env.GetOrDefaultString(EnvEdgeRc, "~/.edgerc")),
		edgegrid.WithSection(env.GetOrDefaultString(EnvEdgeRcSection, "default")),
	)
	if err != nil {
		return nil, fmt.Errorf("edgedns: %w", err)
	}

	conf.MaxBody = maxBody

	accountSwitchKey := env.GetOrDefaultString(EnvAccountSwitchKey, "")

	if accountSwitchKey != "" {
		conf.AccountKey = accountSwitchKey
	}

	config := NewDefaultConfig()
	config.Config = conf

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for EdgeDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("edgedns: the configuration of the DNS provider is nil")
	}

	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("edgedns: %w", err)
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	sess, err := session.New(session.WithSigner(d.config))
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	client := edgegriddns.Client(sess)

	zone, err := getZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	record, err := client.GetRecord(ctx, edgegriddns.GetRecordRequest{
		Zone:       zone,
		Name:       info.EffectiveFQDN,
		RecordType: "TXT",
	})
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("edgedns: %w", err)
	}

	if err == nil && record == nil {
		return errors.New("edgedns: unknown error")
	}

	if record != nil {
		log.Infof("TXT record already exists. Updating target")

		if containsValue(record.Target, info.Value) {
			// have a record and have entry already
			return nil
		}

		record.Target = append(record.Target, `"`+info.Value+`"`)
		record.TTL = d.config.TTL

		err = client.UpdateRecord(ctx, edgegriddns.UpdateRecordRequest{
			Record: &edgegriddns.RecordBody{
				Name:       record.Name,
				RecordType: record.RecordType,
				TTL:        record.TTL,
				Active:     record.Active,
				Target:     record.Target,
			},
			Zone: zone,
		})
		if err != nil {
			return fmt.Errorf("edgedns: %w", err)
		}

		return nil
	}

	err = client.CreateRecord(ctx, edgegriddns.CreateRecordRequest{
		Record: &edgegriddns.RecordBody{
			Name:       info.EffectiveFQDN,
			RecordType: "TXT",
			TTL:        d.config.TTL,
			Target:     []string{`"` + info.Value + `"`},
		},
		Zone:    zone,
		RecLock: nil,
	})
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	return nil
}

// CleanUp removes the record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	sess, err := session.New(session.WithSigner(d.config))
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	client := edgegriddns.Client(sess)

	zone, err := getZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("edgedns: %w", err)
	}

	existingRec, err := client.GetRecord(ctx, edgegriddns.GetRecordRequest{
		Zone:       zone,
		Name:       info.EffectiveFQDN,
		RecordType: "TXT",
	})
	if err != nil {
		if isNotFound(err) {
			return nil
		}

		return fmt.Errorf("edgedns: %w", err)
	}

	if existingRec == nil {
		return errors.New("edgedns: unknown failure")
	}

	if len(existingRec.Target) == 0 {
		return errors.New("edgedns: TXT record is invalid")
	}

	if !containsValue(existingRec.Target, info.Value) {
		return nil
	}

	newRData := filterRData(existingRec, info)

	if len(newRData) > 0 {
		existingRec.Target = newRData

		err = client.UpdateRecord(ctx, edgegriddns.UpdateRecordRequest{
			Record: &edgegriddns.RecordBody{
				Name:       existingRec.Name,
				RecordType: existingRec.RecordType,
				TTL:        existingRec.TTL,
				Active:     existingRec.Active,
				Target:     existingRec.Target,
			},
			Zone: zone,
		})
		if err != nil {
			return fmt.Errorf("edgedns: %w", err)
		}

		return nil
	}

	err = client.DeleteRecord(ctx, edgegriddns.DeleteRecordRequest{
		Zone:       zone,
		Name:       existingRec.Name,
		RecordType: "TXT",
		RecLock:    nil,
	})
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
	return slices.ContainsFunc(values, func(val string) bool {
		return strings.Trim(val, `"`) == value
	})
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	var e *edgegriddns.Error

	return errors.As(err, &e) && e.StatusCode == http.StatusNotFound
}

func filterRData(existingRec *edgegriddns.GetRecordResponse, info dns01.ChallengeInfo) []string {
	var newRData []string

	for _, val := range existingRec.Target {
		val = strings.Trim(val, `"`)
		if val == info.Value {
			continue
		}

		newRData = append(newRData, val)
	}

	return newRData
}
