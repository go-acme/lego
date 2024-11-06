// Package selectelv2 implements a DNS provider for solving the DNS-01 challenge using Selectel Domains APIv2.
package selectelv2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	selectelapi "github.com/selectel/domains-go/pkg/v2"
	"github.com/selectel/go-selvpcclient/v3/selvpcclient"
	"golang.org/x/net/idna"
)

const tokenHeader = "X-Auth-Token"

const (
	defaultBaseURL            = "https://api.selectel.ru/domains/v2"
	defaultTTL                = 60
	defaultPropagationTimeout = 120 * time.Second
	defaultPollingInterval    = 5 * time.Second
	defaultHTTPTimeout        = 30 * time.Second
)

const (
	envNamespace = "SELECTELV2_"

	EnvBaseURL    = envNamespace + "BASE_URL"
	EnvUsernameOS = envNamespace + "USERNAME"
	EnvPasswordOS = envNamespace + "PASSWORD"
	EnvAccount    = envNamespace + "ACCOUNT_ID"
	EnvProjectID  = envNamespace + "PROJECT_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var errNotFound = errors.New("rrset not found")

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	Username           string
	Password           string
	Account            string
	ProjectID          string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            env.GetOrDefaultString(EnvBaseURL, defaultBaseURL),
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, defaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, defaultHTTPTimeout),
		},
	}
}

type DNSProvider struct {
	baseClient selectelapi.DNSClient[selectelapi.Zone, selectelapi.RRSet]
	config     *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Selectel Domains APIv2.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsernameOS, EnvPasswordOS, EnvAccount, EnvProjectID)
	if err != nil {
		return nil, fmt.Errorf("selectelv2: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsernameOS]
	config.Password = values[EnvPasswordOS]
	config.Account = values[EnvAccount]
	config.ProjectID = values[EnvProjectID]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for selectel.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("selectelv2: the configuration of the DNS provider is nil")
	}

	if config.Username == "" {
		return nil, errors.New("selectelv2: missing username")
	}

	if config.Password == "" {
		return nil, errors.New("selectelv2: missing password")
	}

	if config.Account == "" {
		return nil, errors.New("selectelv2: missing account")
	}

	if config.ProjectID == "" {
		return nil, errors.New("selectelv2: missing project ID")
	}

	headers := http.Header{}
	useragent.SetHeader(headers)

	return &DNSProvider{
		baseClient: selectelapi.NewClient(config.BaseURL, config.HTTPClient, headers),
		config:     config,
	}, nil
}

// Timeout returns the Timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (p *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return p.config.PropagationTimeout, p.config.PollingInterval
}

// Present creates a TXT record to fulfill DNS-01 challenge.
func (p *DNSProvider) Present(domain, _, keyAuth string) error {
	ctx := context.Background()

	client, err := p.authorize()
	if err != nil {
		return fmt.Errorf("selectelv2: authorize: %w", err)
	}

	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := client.getZone(ctx, domain)
	if err != nil {
		return fmt.Errorf("selectelv2: get zone: %w", err)
	}

	rrset, err := client.getRRset(ctx, dns01.UnFqdn(info.EffectiveFQDN), zone.ID)
	if err != nil {
		if !errors.Is(err, errNotFound) {
			return fmt.Errorf("selectelv2: get RRSet: %w", err)
		}

		newRRSet := &selectelapi.RRSet{
			Name:    info.EffectiveFQDN,
			Type:    selectelapi.TXT,
			TTL:     p.config.TTL,
			Records: []selectelapi.RecordItem{{Content: fmt.Sprintf("%q", info.Value)}},
		}

		_, err = client.CreateRRSet(ctx, zone.ID, newRRSet)
		if err != nil {
			return fmt.Errorf("selectelv2: create RRSet: %w", err)
		}

		return nil
	}

	rrset.Records = append(rrset.Records, selectelapi.RecordItem{Content: fmt.Sprintf("%q", info.Value)})

	err = client.UpdateRRSet(ctx, zone.ID, rrset.ID, rrset)
	if err != nil {
		return fmt.Errorf("selectelv2: update RRSet: %w", err)
	}

	return nil
}

// CleanUp removes a TXT record used for DNS-01 challenge.
func (p *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	ctx := context.Background()

	client, err := p.authorize()
	if err != nil {
		return fmt.Errorf("selectelv2: authorize: %w", err)
	}

	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := client.getZone(ctx, domain)
	if err != nil {
		return fmt.Errorf("selectelv2: get zone: %w", err)
	}

	rrset, err := client.getRRset(ctx, dns01.UnFqdn(info.EffectiveFQDN), zone.ID)
	if err != nil {
		return fmt.Errorf("selectelv2: get RRSet: %w", err)
	}

	if len(rrset.Records) <= 1 {
		err = client.DeleteRRSet(ctx, zone.ID, rrset.ID)
		if err != nil {
			return fmt.Errorf("selectelv2: %w", err)
		}

		return nil
	}

	for i, item := range rrset.Records {
		if strings.Trim(item.Content, `"`) == info.Value {
			rrset.Records = append(rrset.Records[:i], rrset.Records[i+1:]...)
			break
		}
	}

	err = client.UpdateRRSet(ctx, zone.ID, rrset.ID, rrset)
	if err != nil {
		return fmt.Errorf("selectelv2: update RRSet: %w", err)
	}

	return nil
}

func (p *DNSProvider) authorize() (*clientWrapper, error) {
	token, err := obtainOpenstackToken(p.config)
	if err != nil {
		return nil, err
	}

	extraHeaders := http.Header{}
	extraHeaders.Set(tokenHeader, token)

	return &clientWrapper{
		DNSClient: p.baseClient.WithHeaders(extraHeaders),
	}, nil
}

func obtainOpenstackToken(config *Config) (string, error) {
	vpcClient, err := selvpcclient.NewClient(&selvpcclient.ClientOptions{
		Username:       config.Username,
		Password:       config.Password,
		UserDomainName: config.Account,
		ProjectID:      config.ProjectID,
	})
	if err != nil {
		return "", fmt.Errorf("new VPC client: %w", err)
	}

	return vpcClient.GetXAuthToken(), nil
}

type clientWrapper struct {
	selectelapi.DNSClient[selectelapi.Zone, selectelapi.RRSet]
}

func (w *clientWrapper) getZone(ctx context.Context, name string) (*selectelapi.Zone, error) {
	unicodeName, err := idna.ToUnicode(name)
	if err != nil {
		return nil, fmt.Errorf("to unicode: %w", err)
	}

	params := &map[string]string{"filter": unicodeName}

	zones, err := w.ListZones(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list zone: %w", err)
	}

	for _, zone := range zones.GetItems() {
		if zone.Name == dns01.ToFqdn(unicodeName) {
			return zone, nil
		}
	}

	if len(strings.Split(dns01.UnFqdn(name), ".")) == 1 {
		return nil, fmt.Errorf("zone '%s' for challenge has not been found", name)
	}

	// -1 can not be returned since if no dots present we exit above
	i := strings.Index(name, ".")

	return w.getZone(ctx, name[i+1:])
}

func (w *clientWrapper) getRRset(ctx context.Context, name, zoneID string) (*selectelapi.RRSet, error) {
	unicodeName, err := idna.ToUnicode(name)
	if err != nil {
		return nil, fmt.Errorf("to unicode: %w", err)
	}

	params := &map[string]string{"name": unicodeName, "rrset_types": string(selectelapi.TXT)}

	resp, err := w.ListRRSets(ctx, zoneID, params)
	if err != nil {
		return nil, fmt.Errorf("list rrset: %w", err)
	}

	for _, rrset := range resp.GetItems() {
		if rrset.Name == dns01.ToFqdn(unicodeName) {
			return rrset, nil
		}
	}

	return nil, errNotFound
}
