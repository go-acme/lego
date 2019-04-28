// Package easydns implements a DNS provider for solving the DNS-01 challenge using EasyDNS API.
package easydns

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

const defaultEndpoint = "https://rest.easydns.net"
const dnsRecordType = "TXT"

type zoneRecord struct {
	ID      string `json:"id,omitempty"`
	Domain  string `json:"domain"`
	Host    string `json:"host"`
	TTL     string `json:"ttl"`
	Prio    string `json:"prio"`
	Type    string `json:"type"`
	Rdata   string `json:"rdata"`
	LastMod string `json:"last_mod,omitempty"`
	Revoked int    `json:"revoked,omitempty"`
	NewHost string `json:"new_host,omitempty"`
}

type addRecordResponse struct {
	Msg    string     `json:"msg"`
	Tm     int        `json:"tm"`
	Data   zoneRecord `json:"data"`
	Status int        `json:"status"`
}

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Endpoint           string
	Token              string
	Key                string
	TTL                int
	URL                *url.URL
	HTTPClient         *http.Client
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		Endpoint:           env.GetOrDefaultString("EASYDNS_ENDPOINT", defaultEndpoint),
		PropagationTimeout: env.GetOrDefaultSecond("EASYDNS_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		SequenceInterval:   env.GetOrDefaultSecond("EASYDNS_SEQUENCE_INTERVAL", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("EASYDNS_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		TTL:                env.GetOrDefaultInt("EASYDNS_TTL", dns01.DefaultTTL),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("EASYDNS_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider describes a provider for acme-proxy
type DNSProvider struct {
	config    *Config
	recordIDs map[string]string
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	url, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("easydns: %v", err)
	}

	config.URL = url
	config.Token = os.Getenv("EASYDNS_TOKEN")
	config.Key = os.Getenv("EASYDNS_KEY")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider .
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("easydns: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("easydns: the API token is missing: EASYDNS_TOKEN")
	}

	if config.Key == "" {
		return nil, errors.New("easydns: the API key is missing: EASYDNS_KEY")
	}

	return &DNSProvider{config: config, recordIDs: map[string]string{}}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, challenge := dns01.GetRecord(domain, keyAuth)

	apiHost, apiDomain := splitFqdn(fqdn)
	record := &zoneRecord{
		Domain: apiDomain,
		Host:   apiHost,
		Type:   dnsRecordType,
		Rdata:  challenge,
		TTL:    strconv.Itoa(d.config.TTL),
		Prio:   "0",
	}

	recordID, err := d.addRecord(apiDomain, record)
	if err != nil {
		return fmt.Errorf("easydns: error adding zone record: %v", err)
	}
	key := getMapKey(fqdn, challenge)
	d.recordIDs[key] = recordID

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, challenge := dns01.GetRecord(domain, keyAuth)

	key := getMapKey(fqdn, challenge)
	recordID, exists := d.recordIDs[key]
	if !exists {
		return nil
	}

	_, apiDomain := splitFqdn(fqdn)
	err := d.deleteRecord(apiDomain, recordID)
	defer delete(d.recordIDs, key)
	if err != nil {
		return fmt.Errorf("easydns: %v", err)
	}

	return nil
}

func (d *DNSProvider) addRecord(domain string, record interface{}) (string, error) {
	path := path.Join("/zones/records/add", domain, dnsRecordType)

	response := &addRecordResponse{}
	err := d.executeRequest(http.MethodPut, path, record, response)
	if err != nil {
		return "", err
	}

	recordID := response.Data.ID

	return recordID, nil
}

func (d *DNSProvider) deleteRecord(domain, recordID string) error {
	path := path.Join("/zones/records", domain, recordID)

	return d.executeRequest(http.MethodDelete, path, nil, nil)
}

func (d *DNSProvider) executeRequest(method, path string, requestMsg, responseMsg interface{}) error {
	reqBody := &bytes.Buffer{}
	if requestMsg != nil {
		err := json.NewEncoder(reqBody).Encode(requestMsg)
		if err != nil {
			return err
		}
	}

	endpoint, err := d.config.URL.Parse(path + "?format=json")
	if err != nil {
		return err
	}

	request, err := http.NewRequest(method, endpoint.String(), reqBody)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.SetBasicAuth(d.config.Token, d.config.Key)

	response, err := d.config.HTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("%d: failed to read response body: %v", response.StatusCode, err)
		}

		return fmt.Errorf("%d: request failed: %v", response.StatusCode, string(body))
	}

	if responseMsg != nil {
		return json.NewDecoder(response.Body).Decode(responseMsg)
	}

	return nil
}

func splitFqdn(fqdn string) (host, domain string) {
	parts := dns.SplitDomainName(fqdn)
	length := len(parts)

	host = strings.Join(parts[0:length-2], ".")
	domain = strings.Join(parts[length-2:length], ".")
	return
}

func getMapKey(fqdn, value string) string {
	return fqdn + "|" + value
}
