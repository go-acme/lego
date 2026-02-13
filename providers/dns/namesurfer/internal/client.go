package internal

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	apiKey    string
	apiSecret string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(baseURL, apiKey, apiSecret string) (*Client, error) {
	if apiKey == "" || apiSecret == "" {
		return nil, errors.New("credentials missing")
	}

	if baseURL == "" {
		return nil, errors.New("base URL missing")
	}

	apiEndpoint, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		BaseURL:   apiEndpoint.JoinPath("jsonrpc10"),
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

// AddDNSRecord adds a DNS record.
func (d *Client) AddDNSRecord(ctx context.Context, zoneName, viewName string, record DNSNode) error {
	digest := d.computeDigest(
		zoneName,
		viewName,
		record.Name,
		record.Type,
		strconv.Itoa(record.TTL),
		record.Data,
	)

	// JSON-RPC 1.0 requires positional parameters array
	params := []any{
		d.apiKey, // keyname
		digest,   // digest
		zoneName, // zonename
		viewName, // viewname
		record,   // record (DNSNode)
	}

	resp, err := d.makeAPICall(ctx, "addDNSRecord", params)
	if err != nil {
		return err
	}

	ok, err := parseBoolean(resp)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("addDNSRecord failed")
	}

	return nil
}

// UpdateDNSHost updates a DNS host record.
// Passing an empty newNode removes the oldNode.
func (d *Client) UpdateDNSHost(ctx context.Context, zoneName, viewName string, oldNode DNSNode) error {
	digest := d.computeDigest(zoneName, viewName)

	// JSON-RPC 1.0 requires positional parameters array
	params := []any{
		d.apiKey,
		digest,
		zoneName,
		viewName,
		oldNode,
		DNSNode{}, // empty node = remove old node
	}

	resp, err := d.makeAPICall(ctx, "updateDNSHost", params)
	if err != nil {
		return err
	}

	ok, err := parseBoolean(resp)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("updateDNSHost failed")
	}

	return nil
}

// SearchDNSHosts searches for DNS host records.
func (d *Client) SearchDNSHosts(ctx context.Context, pattern string) ([]DNSNode, error) {
	digest := d.computeDigest(pattern)

	// JSON-RPC 1.0 requires positional parameters array
	params := []any{
		d.apiKey,
		digest,
		pattern,
	}

	resp, err := d.makeAPICall(ctx, "searchDNSHosts", params)
	if err != nil {
		return nil, err
	}

	var nodes []DNSNode

	err = json.Unmarshal(resp, &nodes)

	return nodes, err
}

// ListZoneBasics lists DNS zones.
func (d *Client) ListZoneBasics(ctx context.Context, mode string) ([]MinimalZone, error) {
	digest := d.computeDigest()

	// JSON-RPC 1.0 requires positional parameters array
	params := []any{
		d.apiKey,
		digest,
		mode,
	}

	resp, err := d.makeAPICall(ctx, "listZoneBasics", params)
	if err != nil {
		return nil, err
	}

	var zones []MinimalZone

	err = json.Unmarshal(resp, &zones)

	return zones, err
}

func (d *Client) makeAPICall(ctx context.Context, method string, params any) (json.RawMessage, error) {
	payload := JSONRPCRequest{
		Method: method,
		Params: params,
		ID:     1,
	}

	buf := new(bytes.Buffer)

	err := json.NewEncoder(buf).Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request JSON body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.BaseURL.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(raw))
	}

	var rpcResp JSONRPCResponse

	if err := json.Unmarshal(raw, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON-RPC response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return rpcResp.Result, nil
}

func (d *Client) computeDigest(parts ...string) string {
	params := []string{d.apiKey}
	params = append(params, parts...)
	params = append(params, d.apiSecret)

	mac := hmac.New(sha256.New, []byte(d.apiSecret))
	mac.Write([]byte(strings.Join(params, "&")))

	return hex.EncodeToString(mac.Sum(nil))
}

func parseBoolean(resp json.RawMessage) (bool, error) {
	var ok bool

	if err := json.Unmarshal(resp, &ok); err != nil {
		return false, err
	}

	return ok, nil
}
