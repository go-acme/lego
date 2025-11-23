// Package gigahostno implements a DNS provider for solving the DNS-01 challenge using Gigahost DNS.
package gigahostno

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
	"github.com/go-acme/lego/v4/providers/dns/gigahostno/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"golang.org/x/net/publicsuffix"
)

// Environment variables names.
const (
	envNamespace = "GIGAHOSTNO_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"
	EnvCode     = envNamespace + "2FA_CODE"
	EnvToken    = envNamespace + "TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const minTTL = 60

// Ensure DNSProvider implements the challenge.ProviderTimeout interface.
var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username string
	Password string
	Code     string // Optional 2FA code
	Token    string // Optional pre-generated token (alternative to username/password)

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	// Token caching for authentication
	token   *internal.Token
	muToken sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Gigahost.
// Credentials can be provided in two ways:
//  1. Username/password authentication (recommended for automatic token refresh):
//     GIGAHOSTNO_USERNAME, GIGAHOSTNO_PASSWORD
//     Optionally: GIGAHOSTNO_2FA_CODE for two-factor authentication
//  2. Pre-generated token authentication:
//     GIGAHOSTNO_TOKEN
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.Token = env.GetOrFile(EnvToken) // Optional

	// If token is provided, use token-based auth
	if config.Token != "" {
		return NewDNSProviderConfig(config)
	}

	// Otherwise, use username/password auth
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("gigahostno: %w", err)
	}

	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]
	config.Code = env.GetOrFile(EnvCode) // Optional

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Gigahost.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gigahostno: the configuration of the DNS provider is nil")
	}

	// Validate credentials: either token OR username+password must be provided
	if config.Token == "" && (config.Username == "" || config.Password == "") {
		return nil, errors.New("gigahostno: credentials missing (provide either GIGAHOSTNO_TOKEN or GIGAHOSTNO_USERNAME+GIGAHOSTNO_PASSWORD)")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("gigahostno: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	config.HTTPClient = clientdebug.Wrap(config.HTTPClient)

	// Only create client if using username/password auth
	// For token auth, client is not needed for authentication
	var client *internal.Client

	if config.Token == "" {
		var err error

		client, err = internal.NewClient(config.Username, config.Password, config.Code)
		if err != nil {
			return nil, fmt.Errorf("gigahostno: %w", err)
		}

		client.HTTPClient = config.HTTPClient
	} else {
		// For token-based auth, we still need a client for API calls
		// Use dummy credentials since we won't authenticate
		var err error

		client, err = internal.NewClient("token-auth", "token-auth", "")
		if err != nil {
			return nil, fmt.Errorf("gigahostno: %w", err)
		}

		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)
	ctx := context.Background()

	return d.withRetryOnAuthError(ctx, func(authToken string) error {
		// Find the zone for this domain
		zone, err := d.findZone(ctx, authToken, dns01.UnFqdn(info.EffectiveFQDN))
		if err != nil {
			return fmt.Errorf("gigahostno: %w", err)
		}

		// Extract subdomain from FQDN
		subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone.Name)
		if err != nil {
			return fmt.Errorf("gigahostno: %w", err)
		}

		// Create the TXT record
		record := internal.CreateRecordRequest{
			Name:  subDomain,
			Type:  "TXT",
			Value: info.Value,
			TTL:   d.config.TTL,
		}

		err = d.client.CreateRecord(ctx, authToken, zone.ID, record)
		if err != nil {
			return fmt.Errorf("gigahostno: failed to create TXT record: %w", err)
		}

		return nil
	})
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)
	ctx := context.Background()

	return d.withRetryOnAuthError(ctx, func(authToken string) error {
		// Find the zone for this domain
		zone, err := d.findZone(ctx, authToken, dns01.UnFqdn(info.EffectiveFQDN))
		if err != nil {
			return fmt.Errorf("gigahostno: %w", err)
		}

		// Extract subdomain from FQDN
		subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone.Name)
		if err != nil {
			return fmt.Errorf("gigahostno: %w", err)
		}

		// Get all records to find the one to delete
		records, err := d.client.ListRecords(ctx, authToken, zone.ID)
		if err != nil {
			return fmt.Errorf("gigahostno: failed to list records: %w", err)
		}

		// Find the TXT record we created
		var recordID string

		for _, r := range records {
			if r.Name == subDomain && r.Type == "TXT" && r.Value == info.Value {
				recordID = r.ID
				break
			}
		}

		if recordID == "" {
			return fmt.Errorf("gigahostno: could not find TXT record: zone=%s, subdomain=%s", zone.ID, subDomain)
		}

		// Delete the record
		err = d.client.DeleteRecord(ctx, authToken, zone.ID, recordID, subDomain, "TXT")
		if err != nil {
			return fmt.Errorf("gigahostno: failed to delete TXT record: %w", err)
		}

		return nil
	})
}

// withRetryOnAuthError executes fn with an auth token, retrying once if auth fails.
func (d *DNSProvider) withRetryOnAuthError(ctx context.Context, fn func(authToken string) error) error {
	authToken, err := d.getAuthToken(ctx)
	if err != nil {
		return fmt.Errorf("gigahostno: authentication failed: %w", err)
	}

	err = fn(authToken)
	if err != nil && isAuthError(err) {
		// Token may have been invalidated server-side, retry once with fresh auth
		d.invalidateToken()

		authToken, err = d.getAuthToken(ctx)
		if err != nil {
			return fmt.Errorf("gigahostno: re-authentication failed: %w", err)
		}

		err = fn(authToken)
	}

	return err
}

// getAuthToken obtains and caches an authentication token.
func (d *DNSProvider) getAuthToken(ctx context.Context) (string, error) {
	// If a pre-generated token was provided, use it directly without caching
	if d.config.Token != "" {
		return d.config.Token, nil
	}

	d.muToken.Lock()
	defer d.muToken.Unlock()

	// Check if we have a valid cached token (with 5 minute buffer)
	if d.token != nil && time.Now().Add(5*time.Minute).Before(d.token.Deadline) {
		return d.token.Token, nil
	}

	// Token is expired/expiring soon or doesn't exist, authenticate
	token, err := d.client.Authenticate(ctx)
	if err != nil {
		return "", err
	}

	d.token = token

	return token.Token, nil
}

// invalidateToken invalidates the cached token, forcing re-authentication on next use.
func (d *DNSProvider) invalidateToken() {
	d.muToken.Lock()
	defer d.muToken.Unlock()

	d.token = nil
}

// isAuthError checks if an error is an authentication error (401 or 403).
func isAuthError(err error) bool {
	var statusCodeErr *errutils.UnexpectedStatusCodeError
	if errors.As(err, &statusCodeErr) {
		return statusCodeErr.StatusCode == http.StatusUnauthorized || statusCodeErr.StatusCode == http.StatusForbidden
	}

	return false
}

// findZone locates the DNS zone for a given domain.
func (d *DNSProvider) findZone(ctx context.Context, token, domain string) (*internal.Zone, error) {
	zones, err := d.client.ListZones(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to list zones: %w", err)
	}

	zone := findBestZone(zones, domain)
	if zone == nil {
		return nil, fmt.Errorf("could not find zone for domain: %s", domain)
	}

	return zone, nil
}

// findBestZone finds the most specific zone matching the domain.
func findBestZone(zones []internal.Zone, domain string) *internal.Zone {
	possibleZones := getPossibleZones(domain)

	var bestMatch *internal.Zone

	var longestMatch int

	for i := range zones {
		zone := &zones[i]

		// Only consider active zones
		if zone.Active != "1" {
			continue
		}

		for _, possible := range possibleZones {
			if zone.Name == possible && len(zone.Name) > longestMatch {
				longestMatch = len(zone.Name)
				bestMatch = zone
			}
		}
	}

	return bestMatch
}

// getPossibleZones returns all possible parent zones for a domain.
func getPossibleZones(domain string) []string {
	var zones []string

	tld, _ := publicsuffix.PublicSuffix(domain)

	for d := range dns01.DomainsSeq(domain) {
		if tld == d {
			// Skip the TLD itself
			break
		}

		zones = append(zones, dns01.UnFqdn(d))
	}

	return zones
}
