package excedo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

const (
	envNamespace = "EXCEDO_"

	envAPIKey             = envNamespace + "API_KEY"
	envAPIURL             = envNamespace + "API_URL"
	envTTL                = envNamespace + "TTL"
	envPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	envPollingInterval    = envNamespace + "POLLING_INTERVAL"

	defaultTTL                = 60
	defaultPropagationTimeout = 5 * time.Minute
	defaultPollingInterval    = 10 * time.Second
)

// Config is used to configure the DNS provider.
type Config struct {
	APIKey             string
	APIURL             string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a config with default values.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                defaultTTL,
		PropagationTimeout: defaultPropagationTimeout,
		PollingInterval:    defaultPollingInterval,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config

	mu           sync.Mutex
	token        string
	tokenExpires time.Time
}

// NewDNSProvider returns a DNSProvider instance configured from environment variables.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	values, err := env.Get(envAPIKey, envAPIURL)
	if err != nil {
		return nil, fmt.Errorf("excedo: %w", err)
	}
	config.APIKey = values[envAPIKey]
	config.APIURL = strings.TrimSpace(values[envAPIURL])

	config.TTL = env.GetOrDefaultInt(envTTL, defaultTTL)
	config.PropagationTimeout = env.GetOrDefaultSecond(envPropagationTimeout, defaultPropagationTimeout)
	config.PollingInterval = env.GetOrDefaultSecond(envPollingInterval, defaultPollingInterval)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig returns a DNSProvider instance configured for the given config.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("excedo: config must not be nil")
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("excedo: missing API key")
	}
	if config.APIURL == "" {
		return nil, fmt.Errorf("excedo: missing API URL")
	}
	if config.HTTPClient == nil {
		return nil, fmt.Errorf("excedo: HTTP client must not be nil")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (p *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("excedo: find zone: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("excedo: extract subdomain: %w", err)
	}

	if subDomain == "" {
		subDomain = dns01.UnFqdn(zone)
	}

	payload := url.Values{
		"domainname": {dns01.UnFqdn(zone)},
		"name":       {subDomain},
		"type":       {"TXT"},
		"content":    {info.Value},
		"ttl":        {fmt.Sprintf("%d", p.config.TTL)},
	}

	_, err = p.request(context.Background(), http.MethodPost, "/dns/addrecord/", nil, payload, nil)
	if err != nil {
		return fmt.Errorf("excedo: add record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record created for the dns-01 challenge.
func (p *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("excedo: find zone: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("excedo: extract subdomain: %w", err)
	}
	if subDomain == "" {
		subDomain = dns01.UnFqdn(zone)
	}

	recordID, err := p.findRecordID(context.Background(), dns01.UnFqdn(zone), subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("excedo: find record: %w", err)
	}
	if recordID == "" {
		return nil
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("domainname", dns01.UnFqdn(zone))
	_ = writer.WriteField("recordid", recordID)
	if err := writer.Close(); err != nil {
		return fmt.Errorf("excedo: build delete payload: %w", err)
	}

	_, err = p.request(context.Background(), http.MethodPost, "/dns/deleterecord/", nil, nil, &multipartPayload{
		Body:        body,
		ContentType: writer.FormDataContentType(),
	})
	if err != nil {
		return fmt.Errorf("excedo: delete record: %w", err)
	}

	return nil
}

// Timeout returns the DNS propagation timeout and polling interval.
func (p *DNSProvider) Timeout() (time.Duration, time.Duration) {
	return p.config.PropagationTimeout, p.config.PollingInterval
}

func (p *DNSProvider) request(ctx context.Context, method, path string, query url.Values, form url.Values, multipartBody *multipartPayload) (*apiResponse, error) {
	req, err := p.newRequest(ctx, method, path, query, form, multipartBody, false)
	if err != nil {
		return nil, err
	}

	resp, err := p.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	payload, err := decodeResponse(resp)
	if err != nil {
		return nil, err
	}
	if payload.Code == 2200 {
		if err := p.login(ctx); err != nil {
			return nil, err
		}
		req, err = p.newRequest(ctx, method, path, query, form, multipartBody, false)
		if err != nil {
			return nil, err
		}
		resp, err = p.config.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		payload, err = decodeResponse(resp)
		if err != nil {
			return nil, err
		}
	}

	if payload.Code != 1000 && payload.Code != 1300 {
		desc := errorDescription(payload.Code)
		if desc == "" {
			desc = payload.Desc
		}
		return nil, fmt.Errorf("excedo: api error %d: %s", payload.Code, desc)
	}

	return payload, nil
}

func (p *DNSProvider) newRequest(ctx context.Context, method, path string, query url.Values, form url.Values, multipartBody *multipartPayload, useAPIKey bool) (*http.Request, error) {
	endpoint := strings.TrimRight(p.config.APIURL, "/") + path
	if len(query) > 0 {
		endpoint = endpoint + "?" + query.Encode()
	}
	var body *bytes.Reader
	headers := make(http.Header)

	switch {
	case multipartBody != nil:
		headers.Set("Content-Type", multipartBody.ContentType)
		body = bytes.NewReader(multipartBody.Body.Bytes())
	case form != nil:
		encoded := form.Encode()
		headers.Set("Content-Type", "application/x-www-form-urlencoded")
		body = bytes.NewReader([]byte(encoded))
	default:
		body = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	token, err := p.getToken(ctx, useAPIKey)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	return req, nil
}

func (p *DNSProvider) login(ctx context.Context) error {
	req, err := p.newRequest(ctx, http.MethodGet, "/authenticate/login/", nil, nil, nil, true)
	if err != nil {
		return err
	}

	resp, err := p.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	payload, err := decodeResponse(resp)
	if err != nil {
		return err
	}

	if payload.Code != 1000 && payload.Code != 1300 {
		return fmt.Errorf("excedo: login failed: %s", payload.Desc)
	}

	if payload.Parameters.Token == "" {
		return fmt.Errorf("excedo: login returned empty token")
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.token = payload.Parameters.Token
	p.tokenExpires = time.Now().Add(2*time.Hour - time.Minute)

	return nil
}

func (p *DNSProvider) getToken(ctx context.Context, useAPIKey bool) (string, error) {
	if useAPIKey {
		return p.config.APIKey, nil
	}

	p.mu.Lock()
	token := p.token
	expires := p.tokenExpires
	p.mu.Unlock()

	if token == "" || time.Now().After(expires) {
		if err := p.login(ctx); err != nil {
			return "", err
		}
		p.mu.Lock()
		token = p.token
		p.mu.Unlock()
	}

	return token, nil
}

func (p *DNSProvider) findRecordID(ctx context.Context, zone, name, value string) (string, error) {
	query := make(url.Values)
	query.Set("domainname", zone)
	payload, err := p.request(ctx, http.MethodGet, "/dns/getrecords/", query, nil, nil)
	if err != nil {
		return "", err
	}

	zoneData, ok := payload.DNS[zone]
	if !ok {
		return "", nil
	}

	var zoneRecords dnsZone
	if err := json.Unmarshal(zoneData, &zoneRecords); err != nil {
		return "", err
	}

	records := zoneRecords.Records
	for _, record := range records {
		if record.Type != "TXT" {
			continue
		}
		if !recordNameMatches(zone, name, record.Name) {
			continue
		}
		if normalizeTXT(record.Content) != normalizeTXT(value) {
			continue
		}
		return record.RecordID, nil
	}

	return "", nil
}

type multipartPayload struct {
	Body        *bytes.Buffer
	ContentType string
}

type apiResponse struct {
	Code       int    `json:"code"`
	Desc       string `json:"desc"`
	Parameters struct {
		Token string `json:"token"`
	} `json:"parameters"`
	DNS dnsPayload `json:"dns"`
}

type record struct {
	RecordID string `json:"recordid"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      string `json:"ttl"`
	Prio     string `json:"prio"`
}

type dnsZone struct {
	Records []record `json:"records"`
}

type dnsPayload map[string]json.RawMessage

func (d *dnsPayload) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*d = nil
		return nil
	}

	if trimmed[0] == '[' {
		var ignored []any
		if err := json.Unmarshal(trimmed, &ignored); err != nil {
			return err
		}
		*d = dnsPayload{}
		return nil
	}

	var zones map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &zones); err != nil {
		return err
	}
	*d = zones

	return nil
}

func decodeResponse(resp *http.Response) (*apiResponse, error) {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("excedo: http status %d", resp.StatusCode)
	}

	var payload apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func normalizeTXT(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.Trim(trimmed, "\"")
	return trimmed
}

func recordNameMatches(zone, challengeName, recordName string) bool {
	zone = strings.ToLower(strings.TrimSpace(dns01.UnFqdn(zone)))
	challengeName = strings.ToLower(strings.TrimSpace(dns01.UnFqdn(challengeName)))
	recordName = strings.ToLower(strings.TrimSpace(dns01.UnFqdn(recordName)))

	if recordName == challengeName {
		return true
	}

	fqdnChallenge := challengeName
	if challengeName != "" && zone != "" && !strings.HasSuffix(challengeName, "."+zone) && challengeName != zone {
		fqdnChallenge = challengeName + "." + zone
	}

	return recordName == fqdnChallenge
}

func errorDescription(code int) string {
	switch code {
	case 1000:
		return "Command completed successfully"
	case 1300:
		return "Command completed successfully; no messages"
	case 2001:
		return "Command syntax error"
	case 2002:
		return "Command use error"
	case 2003:
		return "Required parameter missing"
	case 2004:
		return "Parameter value range error"
	case 2104:
		return "Billing failure"
	case 2200:
		return "Authentication error"
	case 2201:
		return "Authorization error"
	case 2303:
		return "Object does not exist"
	case 2304:
		return "Object status prohibits operation"
	case 2309:
		return "Object duplicate found"
	case 2400:
		return "Command failed"
	case 2500:
		return "Command failed; server closing connection"
	default:
		return ""
	}
}
