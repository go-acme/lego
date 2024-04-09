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
	"github.com/go-acme/lego/v4/providers/dns/internal/selectel"
	selectelapi "github.com/selectel/domains-go/pkg/v2"
	"github.com/selectel/go-selvpcclient/v3/selvpcclient"
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

var RRsetNotFoundErr = errors.New("rrset for challenge has not been found")

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
		BaseURL:            env.GetOrDefaultString(EnvBaseURL, selectel.DefaultSelectelBaseURL),
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, defaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, defaultHTTPTimeout),
		},
	}
}

type DNSProvider struct {
	client      selectelapi.DNSClient[selectelapi.Zone, selectelapi.RRSet]
	providerCtx context.Context
	config      *Config
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
	headers := http.Header{}
	headers.Set("User-Agent", "lego/selectelv2")

	return &DNSProvider{
		client:      selectelapi.NewClient(defaultBaseURL, config.HTTPClient, headers),
		providerCtx: context.Background(),
		config:      config,
	}, nil
}

// Timeout returns the Timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (p *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return p.config.PropagationTimeout, p.config.PollingInterval
}

// Present creates a TXT record to fulfill DNS-01 challenge.
func (p *DNSProvider) Present(domain, _, keyAuth string) error {
	err := p.authorize()
	if err != nil {
		return fmt.Errorf("selectelv2: %w", err)
	}

	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := p.getZoneByName(domain)
	if err != nil {
		return fmt.Errorf("selectelv2: %w", err)
	}

	value := fmt.Sprintf("%q", info.Value)

	rrset, err := p.getChallengeRRset(dns01.UnFqdn(info.EffectiveFQDN), zone.ID)
	if err != nil {
		if !errors.Is(err, RRsetNotFoundErr) {
			return err
		}

		newRRSet := &selectelapi.RRSet{
			Name:    info.EffectiveFQDN,
			Type:    selectelapi.TXT,
			TTL:     p.config.TTL,
			Records: []selectelapi.RecordItem{{Content: value}},
		}

		_, err = p.client.CreateRRSet(p.providerCtx, zone.ID, newRRSet)
		if err != nil {
			return fmt.Errorf("selectelv2: %w", err)
		}

		return nil
	}

	rrset.Records = append(rrset.Records, selectelapi.RecordItem{Content: value})

	err = p.client.UpdateRRSet(p.providerCtx, zone.ID, rrset.ID, rrset)
	if err != nil {
		return fmt.Errorf("selectelv2: %w", err)
	}

	return nil
}

// CleanUp removes a TXT record used for DNS-01 challenge.
func (p *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	err := p.authorize()
	if err != nil {
		return fmt.Errorf("selectelv2: %w", err)
	}

	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := p.getZoneByName(domain)
	if err != nil {
		return fmt.Errorf("selectelv2: %w", err)
	}

	rrset, err := p.getChallengeRRset(dns01.UnFqdn(info.EffectiveFQDN), zone.ID)
	if err != nil {
		return fmt.Errorf("selectelv2: %w", err)
	}

	if len(rrset.Records) <= 1 {
		err = p.client.DeleteRRSet(p.providerCtx, zone.ID, rrset.ID)
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

	err = p.client.UpdateRRSet(p.providerCtx, zone.ID, rrset.ID, rrset)
	if err != nil {
		return fmt.Errorf("selectelv2: %w", err)
	}

	return nil
}

func (p *DNSProvider) getZoneByName(name string) (*selectelapi.Zone, error) {
	params := &map[string]string{"filter": name}

	l, err := p.client.ListZones(p.providerCtx, params)
	if err != nil {
		return nil, fmt.Errorf("find zone: %w", err)
	}

	for _, z := range l.GetItems() {
		if z.Name == dns01.ToFqdn(name) {
			return z, nil
		}
	}

	if len(strings.Split(strings.TrimRight(name, "."), ".")) == 1 {
		return nil, errors.New("zone for challenge has not been found")
	}

	// -1 can not be returned since if no dots present we exit above
	i := strings.Index(name, ".")

	return p.getZoneByName(name[i+1:])
}

func (p *DNSProvider) getChallengeRRset(name, zoneID string) (*selectelapi.RRSet, error) {
	params := &map[string]string{"name": name, "rrset_types": string(selectelapi.TXT)}

	resp, err := p.client.ListRRSets(p.providerCtx, zoneID, params)
	if err != nil {
		return nil, fmt.Errorf("find rrset: %w", err)
	}

	for _, rrset := range resp.GetItems() {
		if rrset.Name == dns01.ToFqdn(name) {
			return rrset, nil
		}
	}

	return nil, RRsetNotFoundErr
}

func (p *DNSProvider) authorize() error {
	key := fmt.Sprintf("_%s_%s", p.config.Username, p.config.ProjectID)

	token := p.providerCtx.Value(key)
	if token != nil {
		extraHeaders := http.Header{}
		extraHeaders.Add(tokenHeader, token.(string))
		p.client = p.client.WithHeaders(extraHeaders)
		return nil
	}

	newToken, err := obtainOpenstackToken(p.config)
	if err != nil {
		return err
	}

	p.providerCtx = context.WithValue(p.providerCtx, key, newToken)

	return p.authorize()
}

func obtainOpenstackToken(config *Config) (string, error) {
	vpcClient, err := selvpcclient.NewClient(&selvpcclient.ClientOptions{
		Username:       config.Username,
		Password:       config.Password,
		UserDomainName: config.Account,
		ProjectID:      config.ProjectID,
	})
	if err != nil {
		return "", fmt.Errorf("authorize: %w", err)
	}

	return vpcClient.GetXAuthToken(), nil
}
