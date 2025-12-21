// Package ispconfig implements a DNS provider for solving the DNS-01 challenge using ISPConfig.
package ispconfig

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/ispconfig/internal"
)

// Environment variables names.
const (
	envNamespace = "ISPCONFIG_"

	EnvServerURL = envNamespace + "SERVER_URL"
	EnvUsername  = envNamespace + "USERNAME"
	EnvPassword  = envNamespace + "PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvInsecureSkipVerify = envNamespace + "INSECURE_SKIP_VERIFY"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ServerURL string
	Username  string
	Password  string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
	InsecureSkipVerify bool
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
	config *Config
	client *internal.Client

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for ISPConfig.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvServerURL, EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("ispconfig: %w", err)
	}

	config := NewDefaultConfig()
	config.ServerURL = values[EnvServerURL]
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]
	config.InsecureSkipVerify = env.GetOrDefaultBool(EnvInsecureSkipVerify, false)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ISPConfig.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ispconfig: the configuration of the DNS provider is nil")
	}

	if config.ServerURL == "" {
		return nil, errors.New("ispconfig: missing server URL")
	}

	if config.Username == "" || config.Password == "" {
		return nil, errors.New("ispconfig: credentials missing")
	}

	client, err := internal.NewClient(config.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("ispconfig: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	if config.InsecureSkipVerify {
		client.HTTPClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]string),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	sessionID, err := d.client.Login(ctx, d.config.Username, d.config.Password)
	if err != nil {
		return fmt.Errorf("ispconfig: login: %w", err)
	}

	zoneID, err := d.findZone(ctx, sessionID, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ispconfig: get zone id: %w", err)
	}

	zone, err := d.client.GetZone(ctx, sessionID, strconv.Itoa(zoneID))
	if err != nil {
		return fmt.Errorf("ispconfig: get zone: %w", err)
	}

	clientID, err := d.client.GetClientID(ctx, sessionID, zone.SysUserID)
	if err != nil {
		return fmt.Errorf("ispconfig: get client id: %w", err)
	}

	params := internal.RecordParams{
		ServerID: "serverA",
		Zone:     zone.ID,
		Name:     info.EffectiveFQDN,
		Type:     "txt",
		Data:     info.Value,
		Aux:      "0",
		TTL:      strconv.Itoa(d.config.TTL),
		Active:   "y",
		Stamp:    time.Now().UTC().Format("2006-01-02 15:04:05"),
	}

	recordID, err := d.client.AddTXT(ctx, sessionID, strconv.Itoa(clientID), params)
	if err != nil {
		return fmt.Errorf("ispconfig: add txt record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
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
		return fmt.Errorf("ispconfig: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	sessionID, err := d.client.Login(ctx, d.config.Username, d.config.Password)
	if err != nil {
		return fmt.Errorf("ispconfig: login: %w", err)
	}

	_, err = d.client.DeleteTXT(ctx, sessionID, recordID)
	if err != nil {
		return fmt.Errorf("ispconfig: delete txt record: %w", err)
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

func (d *DNSProvider) findZone(ctx context.Context, sessionID, fqdn string) (int, error) {
	for domain := range dns01.UnFqdnDomainsSeq(fqdn) {
		zoneID, err := d.client.GetZoneID(ctx, sessionID, domain)
		if err == nil {
			return zoneID, nil
		}
	}

	return 0, fmt.Errorf("zone not found for %q", fqdn)
}
