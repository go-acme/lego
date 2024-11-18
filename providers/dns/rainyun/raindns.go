package rainyun

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"io"
	"net/http"
	"strings"
	_ "strings"
	"time"
)

// Environment variables names.
const (
	envNamespace = "RAIN_"

	EnvApiKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *http.Client
}

func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	values, err := env.Get(EnvApiKey)
	if err != nil {
		return nil, fmt.Errorf("rainyun: %w", err)
	}

	config.APIKey = values[EnvApiKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for alidns.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("rainyun: the configuration of the DNS provider is nil")
	}
	c := &http.Client{Timeout: config.HTTPTimeout}
	return &DNSProvider{config: config, client: c}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	//创建TXT record
	domainId, err := d.FindDomainId(domain)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.v2.rainyun.com/product/domain/%d/dns", domainId)
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
    "host": "%s",
    "level": 10,
    "line": "DEFAULT",
    "ttl": 300,
    "type": "TXT",
    "value": "%s"
}`, info.FQDN[:len(info.FQDN)-1], info.Value))

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return fmt.Errorf("rainyun: API call failed: %w", err)
	}
	req.Header.Add("x-api-key", d.config.APIKey)

	res, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("rainyun: API call failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(res.Body)

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	//删除TXT record
	domainId, err := d.FindDomainId(domain)
	if err != nil {
		return err
	}
	recordId, err := d.FindRecordId(domainId, info)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.v2.rainyun.com/product/domain/%d/dns?record_id=%d", domainId, recordId)
	method := "DELETE"

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return fmt.Errorf("rainyun: API call failed: %w", err)
	}
	req.Header.Add("x-api-key", d.config.APIKey)

	res, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("rainyun: API call failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(res.Body)

	return nil
}

type DomainResp struct {
	Code string `json:"name"`
	Data struct {
		TotalRecords int `json:"TotalRecords"`
		Records      []struct {
			Id     int    `json:"id"`
			Domain string `json:"domain"`
		} `json:"Records"`
	} `json:"data"`
}

func (d *DNSProvider) DomainList() (*DomainResp, error) {
	url := "https://api.v2.rainyun.com/product/domain/?options={\"columnFilters\":{\"domains.Domain\":\"\"},\"sort\":[],\"page\":1,\"perPage\":100}"
	method := "GET"

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, fmt.Errorf("rainyun: API call failed: %w", err)
	}

	req.Header.Add("x-api-key", d.config.APIKey)

	res, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rainyun: API call failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("rainyun: API call failed: %w", err)
	}

	// 解析JSON
	var domainData DomainResp
	err = json.Unmarshal(body, &domainData)
	if err != nil {
		return nil, fmt.Errorf("rainyun: API call failed: %w", err)
	}
	return &domainData, nil
}

type RecordResp struct {
	Code string `json:"name"`
	Data struct {
		TotalRecords int `json:"TotalRecords"`
		Records      []struct {
			Id    int    `json:"record_id"`
			Host  string `json:"host"`
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"Records"`
	} `json:"data"`
}

func (d *DNSProvider) RecordList(id int) (*RecordResp, error) {
	url := fmt.Sprintf(`https://api.v2.rainyun.com/product/domain/%d/dns/?limit=100&page_no=1`, id)
	method := "GET"

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, fmt.Errorf("rainyun: API call failed: %w", err)
	}
	req.Header.Add("x-api-key", d.config.APIKey)

	res, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rainyun: API call failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("rainyun: API call failed: %w", err)
	}
	// 解析JSON
	var recordData RecordResp
	err = json.Unmarshal(body, &recordData)
	if err != nil {
		return nil, fmt.Errorf("rainyun: API call failed: %w", err)
	}
	return &recordData, nil
}

func (d *DNSProvider) FindDomainId(domain string) (int, error) {
	var domainId = 0
	list, err := d.DomainList()
	if err != nil {
		return domainId, err
	}

	for _, record := range list.Data.Records {
		if record.Domain == domain {
			domainId = record.Id
			break
		}
	}
	if domainId <= 0 {
		return domainId, fmt.Errorf("rainyun: no domain found for domain %s", domain)
	}
	return domainId, nil
}

func (d *DNSProvider) FindRecordId(id int, info dns01.ChallengeInfo) (int, error) {
	var recordId = 0
	list, err := d.RecordList(id)
	if err != nil {
		return recordId, err
	}
	for _, record := range list.Data.Records {
		if record.Host == info.FQDN[:len(info.FQDN)-1] && record.Value == info.Value {
			recordId = record.Id
			break
		}
	}
	if recordId <= 0 {
		return recordId, fmt.Errorf("rainyun: no record found for record %s", info.FQDN)
	}
	return recordId, nil
}
