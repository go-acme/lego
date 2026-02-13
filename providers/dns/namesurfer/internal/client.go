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

	var ok bool

	err := d.doRequest(ctx, "addDNSRecord", params, &ok)
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

	var ok bool

	err := d.doRequest(ctx, "updateDNSHost", params, &ok)
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

	var nodes []DNSNode

	err := d.doRequest(ctx, "searchDNSHosts", params, &nodes)
	if err != nil {
		return nil, err
	}

	return nodes, nil
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

	var zones []DNSZone

	err := d.doRequest(ctx, "listZones", params, &zones)
	if err != nil {
		return nil, err
	}

	return zones, nil
}

func (d *Client) doRequest(ctx context.Context, method string, params []any, result any) error {
	payload := APIRequest{
		Method: method,
		Params: slices.Concat([]any{d.apiKey}, params),
		ID:     1,
	}

	buf := new(bytes.Buffer)

	err := json.NewEncoder(buf).Encode(payload)
	if err != nil {
		return fmt.Errorf("failed to create request JSON body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.BaseURL.String(), buf)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	var rpcResp APIResponse

	err = json.Unmarshal(raw, &rpcResp)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if rpcResp.Error != nil {
		return rpcResp.Error
	}

	err = json.Unmarshal(rpcResp.Result, result)
	if err != nil {
		return err
	}

	return nil
}

func (d *Client) computeDigest(parts ...string) string {
	params := []string{d.apiKey}
	params = append(params, parts...)
	params = append(params, d.apiSecret)

	mac := hmac.New(sha256.New, []byte(d.apiSecret))
	mac.Write([]byte(strings.Join(params, "&")))

	return hex.EncodeToString(mac.Sum(nil))
}
