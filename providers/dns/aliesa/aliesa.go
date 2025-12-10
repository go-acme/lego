// Package aliesa implements a DNS provider for solving the DNS-01 challenge using AlibabaCloud ESA.
package aliesa

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/aliyun/credentials-go/credentials"
	esa "github.com/go-acme/esa-20240910/v2/client"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/ptr"
)

// Environment variables names.
const (
	envNamespace = "ALIESA_"

	EnvRAMRole       = envNamespace + "RAM_ROLE"
	EnvAccessKey     = envNamespace + "ACCESS_KEY"
	EnvSecretKey     = envNamespace + "SECRET_KEY"
	EnvSecurityToken = envNamespace + "SECURITY_TOKEN"
	EnvRegionID      = envNamespace + "REGION_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const defaultRegionID = "cn-hangzhou"

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	RAMRole       string
	APIKey        string
	SecretKey     string
	SecurityToken string
	RegionID      string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *esa.Client

	recordIDs   map[string]int64
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for AlibabaCloud ESA.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.RegionID = env.GetOrFile(EnvRegionID)

	values, err := env.Get(EnvRAMRole)
	if err == nil {
		config.RAMRole = values[EnvRAMRole]
		return NewDNSProviderConfig(config)
	}

	values, err = env.Get(EnvAccessKey, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("aliesa: %w", err)
	}

	config.APIKey = values[EnvAccessKey]
	config.SecretKey = values[EnvSecretKey]
	config.SecurityToken = env.GetOrFile(EnvSecurityToken)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for AlibabaCloud ESA.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("aliesa: the configuration of the DNS provider is nil")
	}

	if config.RegionID == "" {
		config.RegionID = defaultRegionID
	}

	cfg := new(openapi.Config).
		SetRegionId(config.RegionID).
		SetReadTimeout(int(config.HTTPTimeout.Milliseconds()))

	switch {
	case config.RAMRole != "":
		// https://www.alibabacloud.com/help/en/ecs/user-guide/attach-an-instance-ram-role-to-an-ecs-instance
		credentialsCfg := new(credentials.Config).
			SetType("ecs_ram_role").
			SetRoleName(config.RAMRole)

		credentialClient, err := credentials.NewCredential(credentialsCfg)
		if err != nil {
			return nil, fmt.Errorf("aliesa: new credential: %w", err)
		}

		cfg = cfg.SetCredential(credentialClient)

	case config.APIKey != "" && config.SecretKey != "" && config.SecurityToken != "":
		cfg = cfg.
			SetAccessKeyId(config.APIKey).
			SetAccessKeySecret(config.SecretKey).
			SetSecurityToken(config.SecurityToken)

	case config.APIKey != "" && config.SecretKey != "":
		cfg = cfg.
			SetAccessKeyId(config.APIKey).
			SetAccessKeySecret(config.SecretKey)

	default:
		return nil, errors.New("aliesa: ram role or credentials missing")
	}

	client, err := esa.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("aliesa: new client: %w", err)
	}

	// Workaround to get a regional URL.
	// https://github.com/alibabacloud-go/esa-20240910/blame/7660e3aab2045d4820e4b83427a154efe0c79319/client/client.go#L27
	// The `EndpointRule` is hardcoded with an empty string, so the region is ignored.
	client.Endpoint = nil
	client.EndpointRule = ptr.Pointer("regional")

	client.Endpoint, err = esa.GetEndpoint(client, dara.String("esa"), client.RegionId, client.EndpointRule, client.Network, client.Suffix, client.EndpointMap, client.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("aliesa: get endpoint: %w", err)
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int64),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	siteID, err := d.getSiteID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("aliesa: %w", err)
	}

	crReq := new(esa.CreateRecordRequest).
		SetSiteId(siteID).
		SetType("TXT").
		SetRecordName(dns01.UnFqdn(info.EffectiveFQDN)).
		SetTtl(int32(d.config.TTL)).
		SetData(new(esa.CreateRecordRequestData).SetValue(info.Value))

	// https://www.alibabacloud.com/help/en/edge-security-acceleration/esa/api-esa-2024-09-10-createrecord
	crResp, err := esa.CreateRecordWithContext(ctx, d.client, crReq, &dara.RuntimeOptions{})
	if err != nil {
		return fmt.Errorf("aliesa: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = ptr.Deref(crResp.Body.GetRecordId())
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	// gets the record's unique ID
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("aliesa: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	drReq := new(esa.DeleteRecordRequest).
		SetRecordId(recordID)

	// https://www.alibabacloud.com/help/en/edge-security-acceleration/esa/api-esa-2024-09-10-deleterecord
	_, err := esa.DeleteRecordWithContext(ctx, d.client, drReq, &dara.RuntimeOptions{})
	if err != nil {
		return fmt.Errorf("aliesa: delete record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getSiteID(ctx context.Context, fqdn string) (int64, error) {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return 0, fmt.Errorf("aliesa: could not find zone for domain %q: %w", fqdn, err)
	}

	lsReq := new(esa.ListSitesRequest).
		SetSiteName(dns01.UnFqdn(authZone)).
		SetSiteSearchType("suffix")

	// https://www.alibabacloud.com/help/en/edge-security-acceleration/esa/api-esa-2024-09-10-listsites
	lsResp, err := esa.ListSitesWithContext(ctx, d.client, lsReq, &dara.RuntimeOptions{})
	if err != nil {
		return 0, fmt.Errorf("list sites: %w", err)
	}

	for f := range dns01.UnFqdnDomainsSeq(fqdn) {
		domain := dns01.UnFqdn(f)

		for _, site := range lsResp.Body.GetSites() {
			if ptr.Deref(site.GetSiteName()) == domain {
				return ptr.Deref(site.GetSiteId()), nil
			}
		}
	}

	return 0, fmt.Errorf("site not found (fqdn: %q)", fqdn)
}
