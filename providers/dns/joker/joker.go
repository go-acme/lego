// Package joker implements a DNS provider for solving the DNS-01 challenge using joker.com DMAPI.
package joker

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/log"
	"github.com/go-acme/lego/platform/config/env"
)

const (
	defaultBaseURL = "https://dmapi.joker.com/request/"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Debug              bool
	BaseURL            string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
	AuthSid            string
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		Debug:              env.GetOrDefaultBool("JOKER_DEBUG", false),
		TTL:                env.GetOrDefaultInt("JOKER_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("JOKER_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("JOKER_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("JOKER_HTTP_TIMEOUT", 60*time.Second),
		},
	}
}

// DNSProvider is an implementation of the ChallengeProviderTimeout interface
// that uses Joker's DMAPI to manage TXT records for a domain.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Joker DMAPI.
// Credentials must be passed in the environment variable JOKER_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("JOKER_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("joker: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["JOKER_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Joker DMAPI.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("joker: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("joker: credentials missing")
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present installs a TXT record for the DNS challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {

	zone, relative, value, err := getRecordNames(domain, keyAuth)
	if err != nil {
		return fmt.Errorf("joker: %v", err)
	}
	if d.config.Debug {
		log.Infof("[%s] joker: adding TXT record %q to zone %q with value %q", domain, relative, zone, value)
	}

	response, err := d.jokerLogin()
	if err != nil {
		return formatResponseError(response, err)
	}

	response, err = d.jokerGetZone(zone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	dnsZone := addTxtEntryToZone(response.Body, relative, value, d.config.TTL)

	response, err = d.jokerPutZone(zone, dnsZone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	return nil
}

// CleanUp removes a TXT record used for a previous DNS challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	zone, relative, _, err := getRecordNames(domain, keyAuth)
	if err != nil {
		return fmt.Errorf("joker: %v", err)
	}
	if d.config.Debug {
		log.Infof("[%s] joker: removing entry %q from zone %q", domain, relative, zone)
	}

	response, err := d.jokerLogin()
	if err != nil {
		return formatResponseError(response, err)
	}
	defer func() {
		// Try to logout in case of errors
		_, _ = d.jokerLogout()
	}()

	response, err = d.jokerGetZone(zone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	dnsZone, modified := removeTxtEntryFromZone(response.Body, relative)
	if modified {
		response, err = d.jokerPutZone(zone, dnsZone)
		if err != nil || response.StatusCode != 0 {
			return formatResponseError(response, err)
		}
	}

	response, err = d.jokerLogout()
	if err != nil {
		return formatResponseError(response, err)
	}
	return nil
}

// jokerLogin performs a login to Joker's DMAPI
func (d *DNSProvider) jokerLogin() (*jokerResponse, error) {
	if d.config.AuthSid != "" {
		// already logged in
		return nil, nil
	}
	response, err := d.postRequest("login", url.Values{"api-key": {d.config.APIKey}})
	if err != nil {
		return response, err
	}
	if response == nil {
		return nil, fmt.Errorf("joker: login returned nil response")
	}
	if response.AuthSid == "" {
		return response, fmt.Errorf("joker: login did not return valid Auth-Sid")
	}
	d.config.AuthSid = response.AuthSid
	return response, nil
}

// jokerLogout closes authenticated session with Joker's DMAPI
func (d *DNSProvider) jokerLogout() (*jokerResponse, error) {
	if d.config.AuthSid != "" {
		response, err := d.postRequest("logout", url.Values{})
		if err == nil {
			d.config.AuthSid = ""
		}
		return response, err
	}
	return nil, fmt.Errorf("joker: already logged out")
}

// jokerGetZone returns content of DNS zone for domain
func (d *DNSProvider) jokerGetZone(domain string) (*jokerResponse, error) {
	if d.config.AuthSid == "" {
		return nil, fmt.Errorf("joker get zone: must be logged in")
	}
	response, err := d.postRequest("dns-zone-get", url.Values{"domain": {dns01.UnFqdn(domain)}})
	return response, err
}

// jokerPutZone uploads DNS zone to Joker DMAPI
func (d *DNSProvider) jokerPutZone(domain, zone string) (*jokerResponse, error) {
	if d.config.AuthSid == "" {
		return nil, fmt.Errorf("joker put zone: must be logged in")
	}
	response, err := d.postRequest("dns-zone-put", url.Values{"domain": {dns01.UnFqdn(domain)}, "zone": {strings.TrimSpace(zone)}})
	if d.config.Debug {
		log.Infof("[%s] dns-zone-put response headers:\n%v\n", domain, response.Headers)
	}
	return response, err
}

// postRequest performs actual HTTP request
func (d *DNSProvider) postRequest(cmd string, data url.Values) (*jokerResponse, error) {
	url := d.config.BaseURL + cmd
	if d.config.AuthSid != "" {
		data.Set("auth-sid", d.config.AuthSid)
	}
	if d.config.Debug {
		log.Infof("postRequest:\n\tURL: %q\n\tData: %v", url, data)
	}
	resp, err := d.config.HTTPClient.PostForm(url, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyString := strings.TrimSpace(string(body))
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error %d [%s]: %v", resp.StatusCode, http.StatusText(resp.StatusCode), body)
	}

	return parseJokerResponse(bodyString), nil
}

// Joker DMAPI Response
type jokerResponse struct {
	Headers    url.Values
	Body       string
	StatusCode int
	StatusText string
	AuthSid    string
}

// parseJokerResponse parses HTTP response body
func parseJokerResponse(message string) *jokerResponse {
	r := jokerResponse{Headers: url.Values{}, StatusCode: -1}
	parts := strings.SplitN(message, "\n\n", 2)
	if len(parts) > 0 {
		for _, line := range strings.Split(parts[0], "\n") {
			if strings.TrimSpace(line) != "" {
				var key, val string
				kv := strings.SplitN(line, ":", 2)
				key = strings.TrimSuffix(kv[0], ":")
				if len(kv) == 2 {
					val = strings.TrimSpace(kv[1])
				} else {
					val = ""
				}
				r.Headers.Add(key, val)
				switch key {
				case "Status-Code":
					i, err := strconv.Atoi(val)
					if err == nil {
						r.StatusCode = i
					}
				case "Status-Text":
					r.StatusText = val
				case "Auth-Sid":
					r.AuthSid = val
				}
			}
		}
	}
	if len(parts) > 1 {
		r.Body = parts[1]
	}
	return &r
}

// getRecordNames returns base DNS domain, relative DNS name and value
func getRecordNames(domain, keyAuth string) (zone, relative, value string, err error) {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	zone, err = dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return
	}
	relative = dns01.UnFqdn(strings.TrimSuffix(fqdn, dns01.ToFqdn(zone)))
	return
}

// formatResponseError formats error with optional details from DMAPI response
func formatResponseError(response *jokerResponse, err error) error {
	if response != nil {
		return fmt.Errorf("joker: DMAPI error: %v Response: %v", err, response.Headers)
	}
	return fmt.Errorf("joker: DMAPI error: %v", err)
}

// removeTxtEntryFromZone clean-ups all TXT records with given name
func removeTxtEntryFromZone(zone, relative string) (string, bool) {
	modified := false
	var zoneEntries []string
	prefix := fmt.Sprintf("%s TXT 0 ", relative)
	for _, line := range strings.Split(zone, "\n") {
		if strings.HasPrefix(line, prefix) {
			modified = true
			continue
		}
		zoneEntries = append(zoneEntries, line)
	}
	return strings.TrimSpace(strings.Join(zoneEntries, "\n")), modified
}

// addTxtEntryToZone returns DNS zone with added TXT record
func addTxtEntryToZone(zone, relative, value string, ttl int) string {
	// prefixForOldEntries := fmt.Sprintf("%s TXT 0 ", relative)
	var zoneEntries []string
	for _, line := range strings.Split(zone, "\n") {
		// if strings.HasPrefix(line, prefixForOldEntries) {
		// 	continue
		// }
		zoneEntries = append(zoneEntries, fixTxtLines(line))
	}

	newZoneEntry := fmt.Sprintf("%s TXT 0 %q %d", relative, value, ttl)
	zoneEntries = append(zoneEntries, newZoneEntry)

	return strings.TrimSpace(strings.Join(zoneEntries, "\n"))
}

// Temporary workaround, until it get fixed on API side
func fixTxtLines(line string) (output string) {
	output = line
	fields := strings.Fields(line)
	if len(fields) < 6 || fields[1] != "TXT" {
		return
	}
	if fields[3][0] == '"' && fields[4] == `"` {
		fields[3] = strings.TrimSpace(fields[3]) + `"`
		fields = append(fields[:4], fields[5:]...)
	}
	return strings.Join(fields, " ")
}
