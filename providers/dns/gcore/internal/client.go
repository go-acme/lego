package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	defaultBaseURL = "https://dnsapi.gcorelabs.com"
	tokenHeader    = "APIKey"
	recordType     = "TXT"
)

type (
	// ResponseErr representation.
	ResponseErr string
	// WrongStatusErr representation.
	WrongStatusErr int
	// ClientOpt for constructor of Client.
	ClientOpt func(*Client)
	// Client for DNS API.
	Client struct {
		HTTPClient *http.Client
		BaseURL    string
		Token      string
	}
)

// Error implementation of error contract.
func (r ResponseErr) Error() string {
	return string(r)
}

// Error implementation of error contract.
func (r WrongStatusErr) Error() string {
	return fmt.Sprintf("wrong status = %d", int(r))
}

// Status info from response.
func (r WrongStatusErr) Status() int {
	return int(r)
}

// NewClient constructor of Client.
func NewClient(token string, opts ...ClientOpt) *Client {
	cl := &Client{Token: token, BaseURL: defaultBaseURL, HTTPClient: &http.Client{}}
	for _, op := range opts {
		op(cl)
	}
	return cl
}

// AddTXTRecord to DNS.
func (c *Client) AddTXTRecord(ctx context.Context, fqdn, value string, ttl int) error {
	zone, err := c.findZone(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("find zone: %w", err)
	}
	rrset := strings.TrimRight(fqdn, ".")
	method := http.MethodPost
	resourceRecords := []resourceRecord{{Content: []string{value}}}
	txt, err := c.zoneTxtRecords(ctx, zone, rrset)
	if err == nil && len(txt.ResourceRecords) > 0 {
		method = http.MethodPut
		resourceRecords = append(resourceRecords, txt.ResourceRecords...)
	}
	err = c.request(
		ctx,
		method,
		fmt.Sprintf("v2/zones/%s/%s/%s", zone, rrset, recordType),
		zoneRecord{
			TTL:             ttl,
			ResourceRecords: resourceRecords,
		},
		nil,
	)
	if err != nil {
		return fmt.Errorf("add record request: %w", err)
	}
	return nil
}

// RemoveTXTRecord from DNS.
func (c *Client) RemoveTXTRecord(ctx context.Context, fqdn, _ string) error {
	zone, err := c.findZone(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("find zone: %w", err)
	}
	err = c.request(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("v2/zones/%s/%s/%s", zone, fqdn, recordType),
		nil,
		nil,
	)
	if err != nil {
		// Support DELETE idempotence https://developer.mozilla.org/en-US/docs/Glossary/Idempotent
		if statusErr := new(WrongStatusErr); errors.As(err, statusErr) &&
			statusErr.Status() == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("delete record request: %w", err)
	}
	return nil
}

func (c *Client) findZone(ctx context.Context, fqdn string) (dnsZone string, err error) {
	possibleZones := extractAllZones(fqdn)
	for _, zone := range possibleZones {
		dnsZone, err = c.zone(ctx, zone)
		if err == nil {
			return dnsZone, nil
		}
	}
	return "", fmt.Errorf("zone not found: %w", err)
}

func (c *Client) zoneTxtRecords(ctx context.Context, zone, rrset string) (result zoneRecord, err error) {
	err = c.request(
		ctx,
		http.MethodGet,
		fmt.Sprintf("v2/zones/%s/%s/%s", zone, rrset, recordType),
		nil,
		&result,
	)
	if err != nil {
		return zoneRecord{}, fmt.Errorf("get zone txt records %s -> %s: %w", zone, rrset, err)
	}
	return result, nil
}

func (c *Client) zone(ctx context.Context, zone string) (string, error) {
	response := zoneResponse{}
	err := c.request(
		ctx,
		http.MethodGet,
		fmt.Sprintf("v2/zones/%s", zone),
		nil,
		&response,
	)
	if err != nil {
		return "", fmt.Errorf("get zone %s: %w", zone, err)
	}
	return response.Name, nil
}

func (c *Client) requestURL(path string) string {
	return strings.TrimRight(c.BaseURL, "/") + "/" + strings.TrimLeft(path, "/")
}

func (c *Client) request(ctx context.Context, method, path string,
	bodyParams interface{}, dest interface{}) (err error) {
	var bs []byte
	if bodyParams != nil {
		bs, err = json.Marshal(bodyParams)
		if err != nil {
			return fmt.Errorf("encode bodyParams: %w", err)
		}
	}
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		c.requestURL(path),
		strings.NewReader(string(bs)),
	)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", tokenHeader, c.Token))
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	if res == nil {
		return fmt.Errorf("response: %w", ResponseErr("nil value"))
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("response: %w", WrongStatusErr(res.StatusCode))
	}
	if dest == nil {
		return nil
	}
	err = json.NewDecoder(res.Body).Decode(dest)
	if err != nil {
		return fmt.Errorf("decode body: %w", err)
	}
	return nil
}

func extractAllZones(fqdn string) []string {
	zones := make([]string, 0)
	parts := strings.Split(strings.TrimRight(fqdn, "."), ".")
	if len(parts) < 3 {
		return zones
	}
	for i := 1; i < len(parts)-1; i++ {
		zones = append(zones, strings.Join(parts[i:], "."))
	}
	return zones
}
