// Package liquidweb implements a DNS provider for solving the DNS-01 challenge using Liquid Web.
package liquidweb

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	lw "github.com/liquidweb/liquidweb-go/client"
	"github.com/liquidweb/liquidweb-go/network"
)

const DefaultBaseUrl = "https://api.liquidweb.com"

// Environment variables names.
const (
	EnvPrefix       = "LWAPI_"
	envLegacyPrefix = "LIQUID_WEB_"

	EnvURL      = "URL"
	EnvUsername = "USERNAME"
	EnvPassword = "PASSWORD"

	EnvZone               = "ZONE"
	EnvTTL                = "TTL"
	EnvPropagationTimeout = "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = "POLLING_INTERVAL"
	EnvHTTPTimeout        = "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	Username           string
	Password           string
	Zone               string
	TTL                int
	PollingInterval    time.Duration
	PropagationTimeout time.Duration
	HTTPTimeout        time.Duration
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config      *Config
	client      *lw.API
	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

func getStringEnv(varName, defVal string) string {
	defVal = env.GetOrDefaultString(envLegacyPrefix+varName, defVal)
	defVal = env.GetOrDefaultString(EnvPrefix+varName, defVal)
	return defVal
}

func getIntEnv(varName string, defVal int) int {
	defVal = env.GetOrDefaultInt(envLegacyPrefix+varName, defVal)
	defVal = env.GetOrDefaultInt(EnvPrefix+varName, defVal)
	return defVal
}

func getSecondEnv(varName string, defVal time.Duration) time.Duration {
	defVal = env.GetOrDefaultSecond(envLegacyPrefix+varName, defVal)
	defVal = env.GetOrDefaultSecond(EnvPrefix+varName, defVal)
	return defVal
}

// NewDNSProvider returns a DNSProvider instance configured for Liquid Web.
func NewDNSProvider() (*DNSProvider, error) {
	config := &Config{
		Username:           getStringEnv(EnvUsername, ""),
		Password:           getStringEnv(EnvPassword, ""),
		BaseURL:            getStringEnv(EnvURL, DefaultBaseUrl),
		Zone:               getStringEnv(EnvZone, ""),
		TTL:                getIntEnv(EnvTTL, 300),
		PropagationTimeout: getSecondEnv(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    getSecondEnv(EnvPollingInterval, 2*time.Second),
		HTTPTimeout:        getSecondEnv(EnvHTTPTimeout, 1*time.Minute),
	}

	if config == nil {
		return nil, errors.New("liquidweb: the configuration of the DNS provider is nil")
	}

	switch {
	case config.Username == "" && config.Password == "":
		return nil, errors.New("liquidweb: username and password are missing, set LWAPI_USERNAME and LWAPI_PASSWORD")
	case config.Username == "":
		return nil, errors.New("liquidweb: username is missing, set LWAPI_USERNAME")
	case config.Password == "":
		return nil, errors.New("liquidweb: password is missing, set LWAPI_PASSWORD")
	}

	// Initialize LW client.
	client, err := lw.NewAPI(config.Username, config.Password, config.BaseURL, int(config.HTTPTimeout.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("liquidweb: could not create Liquid Web API client: %w", err)
	}

	return &DNSProvider{
		config:    config,
		recordIDs: make(map[string]int),
		client:    client,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (time.Duration, time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	params := &network.DNSRecordParams{
		Name:  dns01.UnFqdn(info.EffectiveFQDN),
		RData: strconv.Quote(info.Value),
		Type:  "TXT",
		Zone:  d.config.Zone,
		TTL:   d.config.TTL,
	}

	fmt.Printf("%#v\n", params)

	if len(params.Zone) == 0 {
		bestZone, err := d.findZone(params.Name)
		if err == nil {
			params.Zone = bestZone
		} else {
			return fmt.Errorf("zone not specified in environment, could not detect best zone: %w", err)
		}
	}

	dnsEntry, err := d.client.NetworkDNS.Create(params)
	if err != nil {
		return fmt.Errorf("liquidweb: could not create TXT record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = int(dnsEntry.ID)
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("liquidweb: unknown record ID for '%s'", domain)
	}

	params := &network.DNSRecordParams{ID: recordID}
	_, err := d.client.NetworkDNS.Delete(params)
	if err != nil {
		return fmt.Errorf("liquidweb: could not remove TXT record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) findZone(fqdn string) (string, error) {
	fqdn = dns01.UnFqdn(fqdn)
	zones, err := d.client.NetworkDNSZone.ListAll()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve zones for account: %w", err)
	}

	// filter the zones on the account to only ones that match
	for id := 0; id < len(zones.Items); {
		if !strings.HasSuffix(fqdn, zones.Items[id].Name) {
			zones.Items = append(zones.Items[id:], zones.Items[:id]...)
		} else {
			id++
		}
	}

	// filter the zones on the account to only ones that
	sort.Slice(zones.Items, func(i, j int) bool {
		return len(zones.Items[i].Name) < len(zones.Items[j].Name)
	})

	// powerdns _only_ looks for records on the longest matching subdomain zone
	// aka, for test.sub.example.com if sub.example.com exists, it will look there
	// it will not look atexample.com even if it also exists
	return zones.Items[0].Name, nil
}
