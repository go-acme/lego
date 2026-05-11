// Package comlaude implements a DNS provider for solving the DNS-01 challenge using Com Laude.
package comlaude

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/comlaude/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "COMLAUDE_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"
	EnvAPIKey   = envNamespace + "API_KEY"
	EnvGroupID  = envNamespace + "GROUP_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username string
	Password string
	APIKey   string
	GroupID  string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config     *Config
	client     *internal.Client
	identifier *internal.Identifier

	zoneIDs     map[string]string
	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Com Laude.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword, EnvAPIKey, EnvGroupID)
	if err != nil {
		return nil, fmt.Errorf("comlaude: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]
	config.APIKey = values[EnvAPIKey]
	config.GroupID = values[EnvGroupID]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Com Laude.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("comlaude: the configuration of the DNS provider is nil")
	}

	if config.Username == "" || config.Password == "" || config.APIKey == "" {
		return nil, errors.New("comlaude: credentials missing")
	}

	if config.GroupID == "" {
		return nil, errors.New("comlaude: group ID missing")
	}

	identifier := internal.NewIdentifier()

	if config.HTTPClient != nil {
		identifier.HTTPClient = config.HTTPClient
	}

	identifier.HTTPClient = clientdebug.Wrap(identifier.HTTPClient)

	client, err := internal.NewClient()
	if err != nil {
		return nil, fmt.Errorf("comlaude: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:     config,
		client:     client,
		identifier: identifier,
		zoneIDs:    make(map[string]string),
		recordIDs:  make(map[string]string),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	tok, err := d.identifier.APILogin(ctx, d.config.Username, d.config.Password, d.config.APIKey)
	if err != nil {
		return fmt.Errorf("comlaude: API login: %w", err)
	}

	ctxAuth := internal.WithContext(ctx, tok.AccessToken)

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("comlaude: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("comlaude: %w", err)
	}

	dom, err := d.findDomain(ctxAuth, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("comlaude: %w", err)
	}

	if dom.ActiveZone == nil {
		return fmt.Errorf("comlaude: domain (id: %s) has no active zone", dom.ID)
	}

	record := internal.RecordRequest{
		Name:    subDomain,
		Type:    "TXT",
		TTL:     d.config.TTL,
		Content: info.Value,
	}

	recordID, err := d.client.CreateRecord(ctxAuth, d.config.GroupID, dom.ActiveZone.ID, record)
	if err != nil {
		return fmt.Errorf("comlaude: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.zoneIDs[token] = dom.ActiveZone.ID
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordIDsMu.Lock()
	zoneID, zoneOK := d.zoneIDs[token]
	recordID, recordOK := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !zoneOK || !recordOK {
		return fmt.Errorf("comlaude: unknown record ID or zone ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	tok, err := d.identifier.APILogin(ctx, d.config.Username, d.config.Password, d.config.APIKey)
	if err != nil {
		return fmt.Errorf("comlaude: API login: %w", err)
	}

	ctxAuth := internal.WithContext(ctx, tok.AccessToken)

	err = d.client.DeleteRecord(ctxAuth, d.config.GroupID, zoneID, recordID)
	if err != nil {
		return fmt.Errorf("comlaude: delete record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.zoneIDs, token)
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findDomain(ctx context.Context, authZone string) (*internal.Domain, error) {
	domainsResp, err := d.client.GetDomains(ctx, d.config.GroupID, "*."+authZone, nil)
	if err != nil {
		return nil, fmt.Errorf("get domains: %w", err)
	}

	for _, dom := range domainsResp.Data {
		if dom.Name == authZone || dom.IDNName == authZone {
			return &dom, nil
		}
	}

	return nil, errors.New("domain not found")
}
