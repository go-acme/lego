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
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
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
// http://95.128.3.201:8053/API/NSService_10#addDNSRecord
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
		digest,
		zoneName,
		viewName,
		record,
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
// http://95.128.3.201:8053/API/NSService_10#updateDNSHost
func (d *Client) UpdateDNSHost(ctx context.Context, zoneName, viewName string, oldNode, newNode DNSNode) error {
	digest := d.computeDigest(zoneName, viewName)

	// JSON-RPC 1.0 requires positional parameters array
	params := []any{
		digest,
		zoneName,
		viewName,
		oldNode,
		newNode,
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
// http://95.128.3.201:8053/API/NSService_10#searchDNSHosts
func (d *Client) SearchDNSHosts(ctx context.Context, pattern string) ([]DNSNode, error) {
	digest := d.computeDigest(pattern)

	// JSON-RPC 1.0 requires positional parameters array
	params := []any{
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

// ListZones lists DNS zones.
// http://95.128.3.201:8053/API/NSService_10#listZones
func (d *Client) ListZones(ctx context.Context, mode string) ([]DNSZone, error) {
	digest := d.computeDigest()

	// JSON-RPC 1.0 requires positional parameters array
	params := []any{
		digest,
		mode,
	}

	resp, err := d.makeAPICall(ctx, "listZones", params)
	if err != nil {
		return nil, err
	}

	var zones []DNSZone

	err = json.Unmarshal(resp, &zones)

	return zones, err
}

func (d *Client) makeAPICall(ctx context.Context, method string, params []any) (json.RawMessage, error) {
	payload := APIRequest{
		Method: method,
		Params: slices.Concat([]any{d.apiKey}, params),
		ID:     1,
	}

	buf := new(bytes.Buffer)

	err := json.NewEncoder(buf).Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request JSON body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.BaseURL.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	var rpcResp APIResponse

	if err := json.Unmarshal(raw, &rpcResp); err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
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
