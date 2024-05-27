// Package scaleway implements a DNS provider for solving the DNS-01 challenge using Scaleway Domains API.
// Token: https://www.scaleway.com/en/docs/generate-an-api-token/
package scaleway

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	scwdomain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	minTTL                    = 60
	defaultPollingInterval    = 10 * time.Second
	defaultPropagationTimeout = 120 * time.Second
)

// The access key is not used by the Scaleway client.
const dumpAccessKey = "SCWXXXXXXXXXXXXXXXXX"

// Environment variables names.
const (
	envNamespace = "SCALEWAY_"

	EnvAPIToken  = envNamespace + "API_TOKEN"
	EnvProjectID = envNamespace + "PROJECT_ID"

	altEnvNamespace = "SCW_"

	EnvAccessKey = altEnvNamespace + "ACCESS_KEY"
	EnvSecretKey = altEnvNamespace + "SECRET_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ProjectID          string
	Token              string // TODO(ldez) rename to SecretKey in the next major.
	AccessKey          string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		AccessKey:          dumpAccessKey,
		TTL:                env.GetOneWithFallback(EnvTTL, minTTL, strconv.Atoi, altEnvName(EnvTTL)),
		PropagationTimeout: env.GetOneWithFallback(EnvPropagationTimeout, defaultPropagationTimeout, env.ParseSecond, altEnvName(EnvPropagationTimeout)),
		PollingInterval:    env.GetOneWithFallback(EnvPollingInterval, defaultPollingInterval, env.ParseSecond, altEnvName(EnvPollingInterval)),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *scwdomain.API
}

// NewDNSProvider returns a DNSProvider instance configured for Scaleway Domains API.
// Credentials must be passed in the environment variables:
// SCALEWAY_API_TOKEN, SCALEWAY_PROJECT_ID.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.GetWithFallback([]string{EnvSecretKey, EnvAPIToken})
	if err != nil {
		return nil, fmt.Errorf("scaleway: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvSecretKey]
	config.AccessKey = env.GetOrDefaultString(EnvAccessKey, dumpAccessKey)
	config.ProjectID = env.GetOrFile(EnvProjectID)

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

	configuration := []scw.ClientOption{
		scw.WithAuth(config.AccessKey, config.Token),
		scw.WithUserAgent("Scaleway Lego's provider"),
	}

	if config.ProjectID != "" {
		configuration = append(configuration, scw.WithDefaultProjectID(config.ProjectID))
	}

	// Create a Scaleway client
	clientScw, err := scw.NewClient(configuration...)
	if err != nil {
		return nil, fmt.Errorf("scaleway: %w", err)
	}

	return &DNSProvider{config: config, client: scwdomain.NewAPI(clientScw)}, nil
}

// Timeout returns the Timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill DNS-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	records := []*scwdomain.Record{{
		Data:    fmt.Sprintf(`%q`, info.Value),
		Name:    info.EffectiveFQDN,
		TTL:     uint32(d.config.TTL),
		Type:    scwdomain.RecordTypeTXT,
		Comment: scw.StringPtr("used by lego"),
	}}

	req := &scwdomain.UpdateDNSZoneRecordsRequest{
		DNSZone: info.EffectiveFQDN,
		Changes: []*scwdomain.RecordChange{{
			Add: &scwdomain.RecordChangeAdd{Records: records},
		}},
		ReturnAllRecords:        scw.BoolPtr(false),
		DisallowNewZoneCreation: true,
	}

	_, err := d.client.UpdateDNSZoneRecords(req)
	if err != nil {
		return fmt.Errorf("scaleway: %w", err)
	}

	return nil
}

// CleanUp removes a TXT record used for DNS-01 challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	recordIdentifier := &scwdomain.RecordIdentifier{
		Name: info.EffectiveFQDN,
		Type: scwdomain.RecordTypeTXT,
		Data: scw.StringPtr(fmt.Sprintf(`%q`, info.Value)),
	}

	req := &scwdomain.UpdateDNSZoneRecordsRequest{
		DNSZone: info.EffectiveFQDN,
		Changes: []*scwdomain.RecordChange{{
			Delete: &scwdomain.RecordChangeDelete{IDFields: recordIdentifier},
		}},
		ReturnAllRecords:        scw.BoolPtr(false),
		DisallowNewZoneCreation: true,
	}

	_, err := d.client.UpdateDNSZoneRecords(req)
	if err != nil {
		return fmt.Errorf("scaleway: %w", err)
	}

	return nil
}

func altEnvName(v string) string {
	return strings.ReplaceAll(v, envNamespace, altEnvNamespace)
}
