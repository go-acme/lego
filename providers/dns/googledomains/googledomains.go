package googledomains

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-jose/go-jose/v3/json"
)

// API Documentation: https://developers.google.com/domains/acme-dns/reference/rest

// static check on interface implementation
var _ challenge.Provider = &DNSProvider{}

type acmeTxtRecord struct {
	Fqdn       string `json:"fqdn"`
	Digest     string `json:"digest"`
	UpdateTime string `json:"update_time,omitempty"`
}

type rotateChallengesRequest struct {
	AccessToken        string          `json:"access_token"`
	RecordsToAdd       []acmeTxtRecord `json:"records_to_add,omitempty"`
	RecordsToRemove    []acmeTxtRecord `json:"records_to_remove,omitempty"`
	KeepExpiredRecords bool            `json:"keep_expired_records,omitempty"`
}

type acmeChallengeSet struct {
	Record []acmeTxtRecord `json:"record"`
}

const rotateChallengesRequestURL = "https://acmedns.googleapis.com/v1/acmeChallengeSets/%s:rotateChallenges"

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessToken        string
	HTTPClient         *http.Client
	PollingInterval    time.Duration
	PropagationTimeout time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("GOOGLE_DOMAINS_HTTP_TIMEOUT", 30*time.Second),
		},
		AccessToken:        env.GetOrDefaultString("GOOGLE_DOMAINS_ACCESS_TOKEN", ""),
		PropagationTimeout: env.GetOrDefaultSecond("GOOGLE_DOMAINS_PROPAGATION_TIMEOUT", 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("GOOGLE_DOMAINS_POLLING_INTERVAL", 2*time.Second),
	}
}

// NewDNSProvider returns the Google Domains DNS provider with a default configuration.
func NewDNSProvider() (*DNSProvider, error) {
	return NewDNSProviderConfig(NewDefaultConfig())
}

// NewDNSProviderConfig returns the Google Domains DNS provider with the provided config.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	return &DNSProvider{config: config}, nil
}

type DNSProvider struct {
	config *Config
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	rec := getAcmeTxtRecord(domain, token, keyAuth)

	rotateReq := rotateChallengesRequest{
		AccessToken:        d.config.AccessToken,
		RecordsToAdd:       []acmeTxtRecord{rec},
		KeepExpiredRecords: false,
	}

	_, err := d.doRequest(domain, rotateReq)
	if err != nil {
		return err
	}
	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	rec := getAcmeTxtRecord(domain, token, keyAuth)

	rotateReq := rotateChallengesRequest{
		AccessToken:        d.config.AccessToken,
		RecordsToRemove:    []acmeTxtRecord{rec},
		KeepExpiredRecords: false,
	}

	_, err := d.doRequest(domain, rotateReq)
	if err != nil {
		return err
	}
	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) doRequest(domain string, rotateReq rotateChallengesRequest) (acmeChallengeSet, error) {
	acmeChallengeSetResp := acmeChallengeSet{}

	reqJson, err := json.Marshal(rotateReq)
	if err != nil {
		return acmeChallengeSetResp, fmt.Errorf("error marshalling rotate request: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf(rotateChallengesRequestURL, domain), bytes.NewBuffer(reqJson))
	if err != nil {
		return acmeChallengeSetResp, fmt.Errorf("error when http.NewRequest: %w", err)
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return acmeChallengeSetResp, fmt.Errorf("error when sending http request: %w", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&acmeChallengeSetResp)
	if err != nil {
		return acmeChallengeSetResp, fmt.Errorf("unable to decode response from google domains API: %w", err)
	}
	return acmeChallengeSetResp, nil
}

func getAcmeTxtRecord(domain, token, keyAuth string) acmeTxtRecord {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	return acmeTxtRecord{
		Fqdn:   fqdn,
		Digest: value,
	}
}
