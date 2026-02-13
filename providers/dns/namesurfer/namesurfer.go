// Package namesurfer implements a DNS provider for solving the DNS-01 challenge using FusionLayer NameSurfer API.
package namesurfer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "NAMESURFER_"

	EnvAPIEndpoint = envNamespace + "API_ENDPOINT"
	EnvAPIKey      = envNamespace + "API_KEY"
	EnvAPISecret   = envNamespace + "API_SECRET"
	EnvView        = envNamespace + "VIEW"
	EnvTTL         = envNamespace + "TTL"

	EnvPropagationTimeout    = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval       = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout           = envNamespace + "HTTP_TIMEOUT"
	EnvTLSInsecureSkipVerify = envNamespace + "TLS_INSECURE_SKIP_VERIFY"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIEndpoint        string
	APIKey             string
	APISecret          string
	View               string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: env.GetOrDefaultBool(EnvTLSInsecureSkipVerify, false),
		},
	}

	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 2*time.Second),
		HTTPClient: &http.Client{
			Timeout:   env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
			Transport: transport,
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config

	zoneCache []MinimalZone
	cacheMu   sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for FusionLayer.
// Credentials must be passed in the environment variables:
// NAMESURFER_API_ENDPOINT, NAMESURFER_API_KEY, NAMESURFER_API_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIEndpoint, EnvAPIKey, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("namesurfer: %w", err)
	}

	config := NewDefaultConfig()
	config.APIEndpoint = values[EnvAPIEndpoint]
	config.APIKey = values[EnvAPIKey]
	config.APISecret = values[EnvAPISecret]
	config.View = env.GetOrDefaultString(EnvView, "")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for FusionLayer.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("namesurfer: the configuration of the DNS provider is nil")
	}

	if config.APIEndpoint == "" || config.APIKey == "" || config.APISecret == "" {
		return nil, errors.New("namesurfer: incomplete credentials")
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, name, err := d.getZoneAndRecord(info.EffectiveFQDN)
	if err != nil {
		return err
	}

	record := DNSNode{
		Name: name,
		Type: "TXT",
		TTL:  d.config.TTL,
		Data: info.Value,
	}

	return d.addDNSRecord(zone, d.config.View, record)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, name, err := d.getZoneAndRecord(info.EffectiveFQDN)
	if err != nil {
		return err
	}

	fullname := zone
	if name != "" {
		fullname = name + "." + zone
	}

	existing, err := d.searchDNSHosts(fullname)
	if err != nil {
		return err
	}

	for _, node := range existing {
		if node.Type == "TXT" && node.Data == info.Value {
			err = d.updateDNSHost(zone, d.config.View, node)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// getZoneAndRecord finds the best matching zone for a given FQDN using API zone discovery.
func (d *DNSProvider) getZoneAndRecord(fqdn string) (string, string, error) {
	domain := dns01.UnFqdn(fqdn)

	zones, err := d.getZonesCached()
	if err != nil {
		return "", "", err
	}

	var best string

	for _, z := range zones {
		if strings.HasSuffix(domain, z.Name) {
			if len(z.Name) > len(best) {
				best = z.Name
			}
		}
	}

	if best == "" {
		return "", "", fmt.Errorf("no matching zone for %s", domain)
	}

	record := strings.TrimSuffix(domain, "."+best)
	record = strings.TrimSuffix(record, ".")

	return best, record, nil
}

// getZonesCached returns cached zone list or fetches from API.
func (d *DNSProvider) getZonesCached() ([]MinimalZone, error) {
	d.cacheMu.Lock()
	defer d.cacheMu.Unlock()

	if d.zoneCache != nil {
		return d.zoneCache, nil
	}

	zones, err := d.listZoneBasics("forward")
	if err != nil {
		return nil, fmt.Errorf("failed to list zones: %w", err)
	}

	d.zoneCache = zones

	return zones, nil
}

// calculateDigest computes HMAC-SHA256 digest for FusionLayer API authentication.
func (d *DNSProvider) calculateDigest(parts ...string) string {
	// Build message: keyname + parts... + secret
	var sb strings.Builder
	sb.WriteString(d.config.APIKey)

	for _, part := range parts {
		sb.WriteString("&")
		sb.WriteString(part)
	}

	sb.WriteString("&")
	sb.WriteString(d.config.APISecret)

	message := sb.String()

	mac := hmac.New(sha256.New, []byte(d.config.APISecret))
	mac.Write([]byte(message))

	return hex.EncodeToString(mac.Sum(nil))
}

// addDNSRecord adds a DNS record via FusionLayer API.
func (d *DNSProvider) addDNSRecord(zoneName, viewName string, record DNSNode) error {
	// Calculate digest from: zonename, viewname, record.name, record.type, record.ttl, record.data
	digest := d.calculateDigest(
		zoneName,
		viewName,
		record.Name,
		record.Type,
		strconv.Itoa(record.TTL),
		record.Data,
	)

	// JSON-RPC 1.0 requires positional parameters array
	params := []any{
		d.config.APIKey, // keyname
		digest,          // digest
		zoneName,        // zonename
		viewName,        // viewname
		record,          // record (DNSNode)
	}

	return d.callBoolean("addDNSRecord", params)
}

// updateDNSHost updates a DNS host record via FusionLayer API.
// Passing an empty newNode removes the oldNode.
func (d *DNSProvider) updateDNSHost(zoneName, viewName string, oldNode DNSNode) error {
	digest := d.calculateDigest(zoneName, viewName)

	params := []any{
		d.config.APIKey,
		digest,
		zoneName,
		viewName,
		oldNode,
		DNSNode{}, // empty node = remove old node
	}

	return d.callBoolean("updateDNSHost", params)
}

// searchDNSHosts searches for DNS host records via FusionLayer API.
func (d *DNSProvider) searchDNSHosts(pattern string) ([]DNSNode, error) {
	digest := d.calculateDigest(pattern)

	params := []any{
		d.config.APIKey,
		digest,
		pattern,
	}

	resp, err := d.makeAPICall("searchDNSHosts", params)
	if err != nil {
		return nil, err
	}

	var nodes []DNSNode

	err = json.Unmarshal(resp, &nodes)

	return nodes, err
}

// listZoneBasics lists DNS zones via FusionLayer API.
func (d *DNSProvider) listZoneBasics(mode string) ([]MinimalZone, error) {
	digest := d.calculateDigest()

	params := []any{
		d.config.APIKey,
		digest,
		mode,
	}

	resp, err := d.makeAPICall("listZoneBasics", params)
	if err != nil {
		return nil, err
	}

	var zones []MinimalZone

	err = json.Unmarshal(resp, &zones)

	return zones, err
}

// callBoolean makes a JSON-RPC call expecting a boolean result.
func (d *DNSProvider) callBoolean(method string, params any) error {
	resp, err := d.makeAPICall(method, params)
	if err != nil {
		return err
	}

	var ok bool
	if err := json.Unmarshal(resp, &ok); err != nil {
		return errors.New("namesurfer: invalid boolean response")
	}

	if !ok {
		return fmt.Errorf("namesurfer: %s returned false", method)
	}

	return nil
}

// makeAPICall makes a JSON-RPC call to FusionLayer API.
func (d *DNSProvider) makeAPICall(method string, params any) (json.RawMessage, error) {
	reqBody := JSONRPCRequest{
		Method: method,
		Params: params,
		ID:     1,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, d.config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON-RPC response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// API request/response structures

// DNSNode represents a DNS record.
type DNSNode struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
	TTL  int    `json:"ttl"`
}

// MinimalZone represents a DNS zone.
type MinimalZone struct {
	Type string `json:"type"`
	Name string `json:"name"`
	View string `json:"view"`
}

// JSONRPCRequest represents a JSON-RPC 1.0 request.
type JSONRPCRequest struct {
	Method string `json:"method"`
	Params any    `json:"params"`
	ID     any    `json:"id"` // Can be int or string depending on API
}

// JSONRPCResponse represents a JSON-RPC response.
type JSONRPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *JSONRPCError   `json:"error"`
	ID     any             `json:"id"` // Can be int or string depending on API
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    any    `json:"code"` // Can be int or string depending on API
	Message string `json:"message"`
}
