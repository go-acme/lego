// Package bluecat implements a DNS provider for solving the DNS-01 challenge using a self-hosted Bluecat Address Manager.
package bluecat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

const (
	configType = "Configuration"
	viewType   = "View"
	zoneType   = "Zone"
	txtType    = "TXTRecord"
)

// Environment variables names.
const (
	envNamespace = "BLUECAT_"

	EnvServerURL  = envNamespace + "SERVER_URL"
	EnvUserName   = envNamespace + "USER_NAME"
	EnvPassword   = envNamespace + "PASSWORD"
	EnvConfigName = envNamespace + "CONFIG_NAME"
	EnvDNSView    = envNamespace + "DNS_VIEW"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	UserName           string
	Password           string
	ConfigName         string
	DNSView            string
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
	config *Config
	token  string
}

// NewDNSProvider returns a DNSProvider instance configured for Bluecat DNS.
// Credentials must be passed in the environment variables: BLUECAT_SERVER_URL, BLUECAT_USER_NAME and BLUECAT_PASSWORD.
// BLUECAT_SERVER_URL should have the scheme, hostname, and port (if required) of the authoritative Bluecat BAM server.
// The REST endpoint will be appended.
// In addition, the Configuration name and external DNS View Name must be passed in BLUECAT_CONFIG_NAME and BLUECAT_DNS_VIEW.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvServerURL, EnvUserName, EnvPassword, EnvConfigName, EnvDNSView)
	if err != nil {
		return nil, fmt.Errorf("bluecat: %w", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values[EnvServerURL]
	config.UserName = values[EnvUserName]
	config.Password = values[EnvPassword]
	config.ConfigName = values[EnvConfigName]
	config.DNSView = values[EnvDNSView]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Bluecat DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("bluecat: the configuration of the DNS provider is nil")
	}

	if config.BaseURL == "" || config.UserName == "" || config.Password == "" || config.ConfigName == "" || config.DNSView == "" {
		return nil, errors.New("bluecat: credentials missing")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record using the specified parameters
// This will *not* create a subzone to contain the TXT record,
// so make sure the FQDN specified is within an extant zone.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.CreateRecord(domain, token, fqdn, value)
}

// CreateRecord creates a TXT record to fulfill the DNS-01 challenge.
func (d *DNSProvider) CreateRecord(domain, token, fqdn, value string) error {
	err := d.login()
	if err != nil {
		return err
	}

	viewID, err := d.lookupViewID(d.config.DNSView)
	if err != nil {
		return err
	}

	parentZoneID, name, err := d.lookupParentZoneID(viewID, fqdn)
	if err != nil {
		return err
	}

	queryArgs := map[string]string{
		"parentId": strconv.FormatUint(uint64(parentZoneID), 10),
	}

	body := bluecatEntity{
		Name:       name,
		Type:       "TXTRecord",
		Properties: fmt.Sprintf("ttl=%d|absoluteName=%s|txt=%s|", d.config.TTL, fqdn, value),
	}

	resp, err := d.sendRequest(http.MethodPost, "addEntity", body, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	addTxtBytes, _ := ioutil.ReadAll(resp.Body)
	addTxtResp := string(addTxtBytes)
	// addEntity responds only with body text containing the ID of the created record
	_, err = strconv.ParseUint(addTxtResp, 10, 64)
	if err != nil {
		return fmt.Errorf("bluecat: addEntity request failed: %s", addTxtResp)
	}

	err = d.deploy(parentZoneID)
	if err != nil {
		return err
	}

	return d.logout()
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.DeleteRecord(domain, token, fqdn, value)
}

// DeleteRecord removes a creates a TXT record from the provider.
func (d *DNSProvider) DeleteRecord(domain, token, fqdn, value string) error {
	err := d.login()
	if err != nil {
		return err
	}

	viewID, err := d.lookupViewID(d.config.DNSView)
	if err != nil {
		return err
	}

	parentID, name, err := d.lookupParentZoneID(viewID, fqdn)
	if err != nil {
		return err
	}

	queryArgs := map[string]string{
		"parentId": strconv.FormatUint(uint64(parentID), 10),
		"name":     name,
		"type":     txtType,
	}

	resp, err := d.sendRequest(http.MethodGet, "getEntityByName", nil, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var txtRec entityResponse
	err = json.NewDecoder(resp.Body).Decode(&txtRec)
	if err != nil {
		return fmt.Errorf("bluecat: %w", err)
	}
	queryArgs = map[string]string{
		"objectId": strconv.FormatUint(uint64(txtRec.ID), 10),
	}

	resp, err = d.sendRequest(http.MethodDelete, http.MethodDelete, nil, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = d.deploy(parentID)
	if err != nil {
		return err
	}

	return d.logout()
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
