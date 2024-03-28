package selectelv2

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	v2 "github.com/selectel/domains-go/pkg/v2"
	"github.com/selectel/go-selvpcclient/v3/selvpcclient"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	tokenHeader               = "X-Auth-Token"
	defaultBaseURL            = "https://api.selectel.ru/domains/v2"
	defaultTTL                = 60
	defaultPropagationTimeout = 120 * time.Second
	defaultPollingInterval    = 5 * time.Second
	defaultHTTPTimeout        = 30 * time.Second
	envNamespace              = "SELECTELV2_"

	envBaseURL    = envNamespace + "BASE_URL"
	envUsernameOS = envNamespace + "USERNAME"
	envPasswordOS = envNamespace + "PASSWORD"
	envAccount    = envNamespace + "ACCOUNT_ID"
	envProjectId  = envNamespace + "PROJECT_ID"

	envTTL                = envNamespace + "TTL"
	envPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	envPollingInterval    = envNamespace + "POLLING_INTERVAL"
	envHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var (
	UsernameMissingErr = errors.New(fmt.Sprintf("'%s' is required", envUsernameOS))
	PasswordMissingErr = errors.New(fmt.Sprintf("'%s' is required", envPasswordOS))
	AccountMissingErr  = errors.New(fmt.Sprintf("'%s' is required", envAccount))
	ProjectMissingErr  = errors.New(fmt.Sprintf("'%s' is required", envProjectId))
	ZoneNotFoundErr    = errors.New("zone for challenge has not been found")
	RRsetNotFoundErr   = errors.New("rrset for challenge has not been found")
)

type (
	SelectelDNSProvider struct {
		client      v2.DNSClient[v2.Zone, v2.RRSet]
		providerCtx context.Context
		config      Config
	}
	Config struct {
		BaseURL            string
		Username           string
		Password           string
		Account            string
		ProjectID          string
		TTL                int
		PropagationTimeout time.Duration
		PollingInterval    time.Duration
		HTTPTimeout        time.Duration
	}
	SelectelError struct {
		err error
	}
)

func (e *SelectelError) Error() string {
	return fmt.Sprintf("selectel: %v", e.err)
}

func parseConfig() (Config, error) {
	username := os.Getenv(envUsernameOS)
	if username == "" {
		return Config{}, UsernameMissingErr
	}
	password := os.Getenv(envPasswordOS)
	if password == "" {
		return Config{}, PasswordMissingErr
	}
	account := os.Getenv(envAccount)
	if account == "" {
		return Config{}, AccountMissingErr
	}
	projectId := os.Getenv(envProjectId)
	if projectId == "" {
		return Config{}, ProjectMissingErr
	}
	return Config{
		BaseURL:            env.GetOrDefaultString(envBaseURL, defaultBaseURL),
		Username:           username,
		Password:           password,
		Account:            account,
		ProjectID:          projectId,
		TTL:                env.GetOrDefaultInt(envTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(envPropagationTimeout, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(envPollingInterval, defaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(envHTTPTimeout, defaultHTTPTimeout),
	}, nil
}

func (p *SelectelDNSProvider) authorize() error {
	key := fmt.Sprintf("_%s_%s", p.config.Username, p.config.ProjectID)
	token := p.providerCtx.Value(key)
	if token != nil {
		extraHeaders := http.Header{}
		extraHeaders.Add(tokenHeader, token.(string))
		p.client = p.client.WithHeaders(extraHeaders)
		return nil
	}
	newToken, err := obtainOpenstackToken(&p.config)
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

// NewDNSProvider returns a DNSProvider instance configured for Selectel Domains APIv2.
func NewDNSProvider() (*SelectelDNSProvider, error) {
	cfg, err := parseConfig()
	if err != nil {
		return nil, &SelectelError{err}
	}

	httpClient := &http.Client{Timeout: cfg.HTTPTimeout}
	defaultHeaders := http.Header{}
	defaultHeaders.Add("User-Agent", "lego/selectelv2")
	client := v2.NewClient(defaultBaseURL, httpClient, defaultHeaders)
	return &SelectelDNSProvider{client: client, providerCtx: context.Background(), config: cfg}, nil
}

// Present creates a TXT record to fulfill DNS-01 challenge.
func (p *SelectelDNSProvider) Present(domain, _, keyAuth string) error {
	err := p.authorize()
	if err != nil {
		return &SelectelError{err}
	}
	info := dns01.GetChallengeInfo(domain, keyAuth)
	zone, err := p.getZoneByName(domain)
	if err != nil {
		return &SelectelError{err}
	}
	rrset, err := p.getChallengeRRset(dns01.UnFqdn(info.EffectiveFQDN), zone.ID)
	challengeItem := v2.RecordItem{Content: fmt.Sprintf(`"%s"`, info.Value)}
	if errors.Is(err, RRsetNotFoundErr) {
		if _, err = p.client.CreateRRSet(
			p.providerCtx,
			zone.ID,
			&v2.RRSet{
				Name:    info.EffectiveFQDN,
				Type:    v2.TXT,
				TTL:     p.config.TTL,
				Records: []v2.RecordItem{challengeItem},
			},
		); err != nil {
			return &SelectelError{err}
		}
	} else {
		rrset.Records = append(rrset.Records, challengeItem)
		if err = p.client.UpdateRRSet(p.providerCtx, zone.ID, rrset.ID, rrset); err != nil {
			return &SelectelError{err}
		}
	}

	return nil
}

// CleanUp removes a TXT record used for DNS-01 challenge.
func (p *SelectelDNSProvider) CleanUp(domain, _, keyAuth string) error {
	err := p.authorize()
	if err != nil {
		return &SelectelError{err}
	}
	info := dns01.GetChallengeInfo(domain, keyAuth)
	rrsetName := dns01.UnFqdn(info.EffectiveFQDN)
	zone, err := p.getZoneByName(domain)
	if err != nil {
		return &SelectelError{err}
	}
	rrset, err := p.getChallengeRRset(rrsetName, zone.ID)
	if err != nil {
		return &SelectelError{err}
	}
	if len(rrset.Records) <= 1 {
		if err = p.client.DeleteRRSet(p.providerCtx, zone.ID, rrset.ID); err != nil {
			return &SelectelError{err}
		}
	} else {
		for i, item := range rrset.Records {
			if strings.Trim(item.Content, "\"") == info.Value {
				rrset.Records = append(rrset.Records[:i], rrset.Records[i+1:]...)
				break
			}
		}
		if err = p.client.UpdateRRSet(p.providerCtx, zone.ID, rrset.ID, rrset); err != nil {
			return &SelectelError{err}
		}
	}

	return nil
}

// Timeout returns the Timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (p *SelectelDNSProvider) Timeout() (timeout, interval time.Duration) {
	return p.config.PropagationTimeout, p.config.PollingInterval
}

func (p *SelectelDNSProvider) getZoneByName(name string) (*v2.Zone, error) {
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
		return nil, ZoneNotFoundErr
	}
	// -1 can not be returned since if no dots present we exit above
	i := strings.Index(name, ".")
	return p.getZoneByName(name[i+1:])
}

func (p *SelectelDNSProvider) getChallengeRRset(name, zoneID string) (*v2.RRSet, error) {
	params := &map[string]string{"name": name, "rrset_types": string(v2.TXT)}
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
