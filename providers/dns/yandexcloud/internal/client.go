package internal

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/google/go-querystring/query"
)

const (
	// DNSAPIEndpoint API endpoint for DNS operations
	DNSAPIEndpoint = "https://dns.api.cloud.yandex.net/dns/v1"

	// IAMAPIEndpoint API endpoint for IAM operations
	IAMAPIEndpoint = "https://iam.api.cloud.yandex.net/iam/v1"
)

// Client is the API client for Yandex Cloud API
type Client struct {
	iamToken string
	client   *http.Client
	key      Key
	folderID string
	ttl      int
}

// NewClient creates a new Client
func NewClient(iamToken, folderID string, ttl int) (*Client, error) {
	// Validate IAM token structure
	keyBytes, err := base64.StdEncoding.DecodeString(iamToken)
	if err != nil {
		return nil, fmt.Errorf("iam token is malformed: %w", err)
	}
	var key Key
	err = json.Unmarshal(keyBytes, &key)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal key data: %w", err)
	}

	return &Client{
		iamToken: iamToken,
		key:      key,
		client:   &http.Client{Timeout: 30 * time.Second},
		folderID: folderID,
		ttl:      ttl,
	}, nil
}

// DnsZone represents a DNS zone in Yandex Cloud.
type DnsZone struct {
	ID   string `json:"id"`
	Zone string `json:"zone"`
}

// ListDnsZonesResponse represents the response from listing DNS zones.
type ListDnsZonesResponse struct {
	DnsZones []DnsZone `json:"dnsZones"`
}

// RecordSet represents a DNS record set in Yandex Cloud.
type RecordSet struct {
	Name string   `json:"name"`
	Type string   `json:"type"`
	TTL  int64    `json:"ttl"`
	Data []string `json:"data"`
}

// ListZonesOptions provides parameters for ListZones
type ListZonesOptions struct {
	FolderID string `url:"folderId,omitempty"`
}

// ListZones retrieves available zones from yandex cloud.
func (c *Client) ListZones(ctx context.Context) ([]DnsZone, error) {
	accessToken, err := c.getIAMToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get IAM token: %w", err)
	}

	opt := ListZonesOptions{FolderID: c.folderID}
	v, err := query.Values(opt)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/zones?%s", DNSAPIEndpoint, v.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	var result ListDnsZonesResponse
	err = c.doJSONRequest(req, &result)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch dns zones: %w", err)
	}

	return result.DnsZones, nil
}

// GetRecordSetOptions contains the options for GetRecordSet
type GetRecordSetOptions struct {
	Name string `url:"name,omitempty"`
	Type string `url:"type,omitempty"`
}

// recordSetsResponse contains the response structure for record sets
type recordSetsResponse struct {
	RecordSets []RecordSet `json:"recordSets"`
}

// GetRecordSet retrieves a record set from the DNS zone.
func (c *Client) GetRecordSet(ctx context.Context, zoneID, name string) (*RecordSet, error) {
	accessToken, err := c.getIAMToken(ctx)
	if err != nil {
		return nil, err
	}

	opt := GetRecordSetOptions{
		Name: name,
		Type: "TXT",
	}

	v, err := query.Values(opt)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/zones/%s:getRecordSets?%s", DNSAPIEndpoint, zoneID, v.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, body)
	}

	var result recordSetsResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if len(result.RecordSets) == 0 {
		return nil, nil
	}

	return &result.RecordSets[0], nil
}

// UpdateRecordSetsRequest contains the request for updating record sets
type UpdateRecordSetsRequest struct {
	Deletions []RecordSet `json:"deletions"`
	Additions []RecordSet `json:"additions"`
}

// UpdateRecordSets updates DNS record sets
func (c *Client) UpdateRecordSets(ctx context.Context, zoneID string, deletions, additions []RecordSet) error {
	accessToken, err := c.getIAMToken(ctx)
	if err != nil {
		return err
	}

	requestData := UpdateRecordSetsRequest{
		Deletions: deletions,
		Additions: additions,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/zones/%s:upsertRecordSets", DNSAPIEndpoint, zoneID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req)
}

// CreateRecordSet creates a new TXT record with the specified name and value.
func (c *Client) CreateRecordSet(ctx context.Context, zoneID, name, value string) error {
	existingRecord, err := c.GetRecordSet(ctx, zoneID, name)
	if err != nil {
		return err
	}

	var record RecordSet
	var deletions []RecordSet

	if existingRecord != nil {
		record = *existingRecord
		deletions = append(deletions, *existingRecord)

		// Check if value already exists
		for _, data := range record.Data {
			if data == value {
				return nil // Value already exists, nothing to do
			}
		}

		// Append new value
		record.Data = append(record.Data, value)
	} else {
		record = RecordSet{
			Name: name,
			Type: "TXT",
			TTL:  int64(c.ttl),
			Data: []string{value},
		}
	}

	return c.UpdateRecordSets(ctx, zoneID, deletions, []RecordSet{record})
}

// RemoveRecordSetValue removes a specific value from a TXT record.
func (c *Client) RemoveRecordSetValue(ctx context.Context, zoneID, name, value string) error {
	existingRecord, err := c.GetRecordSet(ctx, zoneID, name)
	if err != nil {
		return err
	}

	if existingRecord == nil {
		return nil // Record doesn't exist, nothing to do
	}

	var additions []RecordSet
	var newData []string

	// Filter out the value to remove
	for _, data := range existingRecord.Data {
		if data != value {
			newData = append(newData, data)
		}
	}

	// If we still have data, create a new record
	if len(newData) > 0 {
		newRecord := *existingRecord
		newRecord.Data = newData
		additions = append(additions, newRecord)
	}

	return c.UpdateRecordSets(ctx, zoneID, []RecordSet{*existingRecord}, additions)
}

// doJSONRequest makes an HTTP request and unmarshals the response.
func (c *Client) doJSONRequest(req *http.Request, v interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, body)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// doRequest makes an HTTP request without unmarshalling the response.
func (c *Client) doRequest(req *http.Request) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, body)
	}

	return nil
}
