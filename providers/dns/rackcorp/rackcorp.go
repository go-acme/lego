package rackcorp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/internal/useragent"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/rackcorp/internal"
)

// Environment variables names.
const (
	envNamespace = "RACKCORP_"

	EnvAPIURL    = envNamespace + "API_URL"
	EnvAPIUUID   = envNamespace + "API_UUID"
	EnvAPISecret = envNamespace + "API_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	URL                string
	APIUUID            string
	APISecret          string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		URL:                env.GetOrDefaultString(EnvAPIURL, internal.DefaultURL),
		APIUUID:            env.GetOrDefaultString(EnvAPIUUID, ""),
		APISecret:          env.GetOrDefaultString(EnvAPISecret, ""),
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}
}

type DNSProvider struct {
	config *Config
	client *internal.RCClient
}

// NewDNSProvider returns a DNSProvider instance configured for Rackcorp.
// Credentials must be passed in the two environment variables,
// `RACKCORP_API_UUID` and `RACKCORP_API_SECRET`.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("rackcorp: the configuration of the DNS provider is nil")
	}

	if config.APIUUID == "" || config.APISecret == "" {
		return nil, errors.New("rackcorp: API credentials are missing")
	}

	client := internal.NewRCClient(clientdebug.Wrap(config.HTTPClient),
		config.URL,
		config.APIUUID,
		config.APISecret,
		useragent.Get(),
	)

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zone, err := d.getHostedZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return err
	}

	zone, err = d.client.DNSDomainGet(zone.ID)
	if err != nil {
		return err
	}

	record := internal.FindTXTRecord(zone.Records, info.Prefix)
	if record != nil {
		record.Data = info.Value
		record.TTL = json.Number(strconv.Itoa(d.config.TTL))

		return d.client.DNSRecordUpdate(*record)
	}

	return d.client.DNSRecordCreateTXT(zone.ID, info.Prefix, info.Value, d.config.TTL)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zone, err := d.getHostedZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return err
	}

	zone, err = d.client.DNSDomainGet(zone.ID)
	if err != nil {
		return err
	}

	record := internal.FindTXTRecord(zone.Records, info.Prefix)
	if record != nil {
		return d.client.DNSRecordDelete(record.ID)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getHostedZone(ctx context.Context, fqdn string) (*internal.DNSDomain, error) {
	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, fqdn)
	if err != nil {
		return nil, fmt.Errorf("could not find zone: %w", err)
	}

	authZone = dns01.UnFqdn(authZone)

	domains, err := d.client.DNSDomainGetAll()
	if err != nil {
		return nil, err
	}

	domain := internal.FindDomain(domains, authZone)
	if domain == nil {
		return nil, fmt.Errorf("could not find domain: %s", authZone)
	}

	return domain, nil
}
