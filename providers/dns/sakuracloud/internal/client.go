package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const (
	// DefaultAPIEndpoint API endpoint for SakuraCloud DNS operations
	DefaultAPIEndpoint = "https://secure.sakura.ad.jp/cloud/zone/is1a/api/cloud/1.1"
)

// Client is the API client for SakuraCloud API
type Client struct {
	token       string
	secret      string
	httpClient  *http.Client
	ttl         int
	apiEndpoint string
}

// NewClient creates a new Client
func NewClient(token, secret string, ttl int) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("SakuraCloud token is required")
	}

	if secret == "" {
		return nil, fmt.Errorf("SakuraCloud secret is required")
	}

	return &Client{
		token:       token,
		secret:      secret,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		ttl:         ttl,
		apiEndpoint: DefaultAPIEndpoint,
	}, nil
}

// Zone represents a DNS zone in SakuraCloud.
type Zone struct {
	ID          string `json:"ID,omitempty"`
	Name        string `json:"Name,omitempty"`
	Description string `json:"Description,omitempty"`
	Status      struct {
		Zone string   `json:"Zone,omitempty"`
		NS   []string `json:"NS,omitempty"`
	} `json:"Status,omitempty"`
	ServiceClass string    `json:"ServiceClass,omitempty"`
	Availability string    `json:"Availability,omitempty"`
	CreatedAt    time.Time `json:"CreatedAt,omitempty"`
	ModifiedAt   time.Time `json:"ModifiedAt,omitempty"`
	Settings     struct {
		DNS struct {
			ResourceRecordSets []Record `json:"ResourceRecordSets,omitempty"`
		} `json:"DNS,omitempty"`
	} `json:"Settings,omitempty"`
}

// ZoneResponse represents the response from getting DNS zones.
type ZoneResponse struct {
	CommonServiceItems []Zone `json:"CommonServiceItems"`
	From               int    `json:"From"`
	Count              int    `json:"Count"`
	Total              int    `json:"Total"`
	IsOK               bool   `json:"is_ok"`
}

// ResponseList represents the common response structure for list operations
type ResponseList struct {
	ZoneResponse
}

// Record represents a DNS record in SakuraCloud.
type Record struct {
	ID     string `json:"ID,omitempty"`
	Name   string `json:"Name,omitempty"`
	Type   string `json:"Type,omitempty"`
	RData  string `json:"RData,omitempty"`
	TTL    int    `json:"TTL,omitempty"`
	ZoneID string `json:"-"`
}

// RecordResponse represents the response from getting DNS records.
type RecordResponse struct {
	From               int `json:"From"`
	Count              int `json:"Count"`
	Total              int `json:"Total"`
	CommonServiceItems []struct {
		ID          string `json:"ID"`
		Name        string `json:"Name"`
		Description string `json:"Description"`
		Settings    struct {
			DNS struct {
				ResourceRecordSets []Record `json:"ResourceRecordSets"`
			} `json:"DNS"`
		} `json:"Settings"`
		Status struct {
			Zone string   `json:"Zone"`
			NS   []string `json:"NS"`
		} `json:"Status"`
	} `json:"CommonServiceItems"`
	IsOK bool `json:"is_ok"`
}

// RecordResponseList represents the common response structure for record operations
type RecordResponseList struct {
	RecordResponse
}

// SingleResponse represents the response from an operation on a single resource
type SingleResponse struct {
	IsOK bool `json:"is_ok"`
}

// GetZoneByDomain retrieves a zone by domain name.
func (c *Client) GetZoneByDomain(ctx context.Context, domain string) (*Zone, error) {
	url := fmt.Sprintf("%s/commonserviceitem", c.apiEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.SetBasicAuth(c.token, c.secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	var result ResponseList
	err = c.doJSONRequest(req, &result)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch dns zones: %w", err)
	}

	if !result.IsOK {
		return nil, fmt.Errorf("API request failed")
	}

	for _, zone := range result.CommonServiceItems {
		if zone.Status.Zone == domain {
			return &zone, nil
		}
	}

	// Find parent zone if exact match not found
	var matchedZone *Zone
	matchLength := 0

	for _, zone := range result.CommonServiceItems {
		if len(zone.Status.Zone) > matchLength && isDomainOrSubdomain(domain, zone.Status.Zone) {
			matchedZone = &zone
			matchLength = len(zone.Status.Zone)
		}
	}

	if matchedZone != nil {
		return matchedZone, nil
	}

	return nil, fmt.Errorf("zone for domain %q not found", domain)
}

// GetRecords retrieves records from a zone.
func (c *Client) GetRecords(ctx context.Context, zoneID string) ([]Record, error) {
	url := fmt.Sprintf("%s/commonserviceitem/%s", c.apiEndpoint, zoneID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.SetBasicAuth(c.token, c.secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	var result RecordResponseList
	err = c.doJSONRequest(req, &result)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch dns records: %w", err)
	}

	if !result.IsOK {
		return nil, fmt.Errorf("API request failed")
	}

	// Ensure at least one CommonServiceItem exists
	if len(result.CommonServiceItems) == 0 {
		return []Record{}, nil
	}

	// Get records from the first CommonServiceItem
	records := result.CommonServiceItems[0].Settings.DNS.ResourceRecordSets

	// Set ZoneID for all records
	for i := range records {
		records[i].ZoneID = zoneID
	}

	return records, nil
}

// CreateTXTRecord creates a new TXT record.
func (c *Client) CreateTXTRecord(ctx context.Context, zoneID, name, value string) error {
	// First get all existing records
	records, err := c.GetRecords(ctx, zoneID)
	if err != nil {
		return fmt.Errorf("failed to get existing records: %w", err)
	}

	// Check if record already exists
	for _, record := range records {
		if record.Type == "TXT" && record.Name == name && record.RData == fmt.Sprintf("\"%s\"", value) {
			return nil // Record already exists, nothing to do
		}
	}

	// Add new record to the array
	newRecord := Record{
		Name:   name,
		Type:   "TXT",
		RData:  fmt.Sprintf("\"%s\"", value),
		TTL:    c.ttl,
		ZoneID: zoneID,
	}
	records = append(records, newRecord)

	// Prepare update request with all records
	data := struct {
		CommonServiceItemID string `json:"CommonServiceItemID"`
		Settings            struct {
			DNS struct {
				ResourceRecordSets []Record `json:"ResourceRecordSets"`
			} `json:"DNS"`
		} `json:"Settings"`
	}{
		CommonServiceItemID: zoneID,
	}
	data.Settings.DNS.ResourceRecordSets = records

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	url := fmt.Sprintf("%s/commonserviceitem/%s", c.apiEndpoint, zoneID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.token, c.secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	var response SingleResponse
	err = c.doJSONRequest(req, &response)
	if err != nil {
		return fmt.Errorf("failed to create TXT record: %w", err)
	}

	if !response.IsOK {
		return fmt.Errorf("API request failed")
	}

	return nil
}

// DeleteTXTRecord deletes a TXT record.
func (c *Client) DeleteTXTRecord(ctx context.Context, zoneID, name, value string) error {
	// Get all existing records
	records, err := c.GetRecords(ctx, zoneID)
	if err != nil {
		return fmt.Errorf("failed to get existing records: %w", err)
	}

	// Find and remove the record
	found := false
	var updatedRecords []Record

	for _, record := range records {
		if record.Type == "TXT" && record.Name == name && record.RData == fmt.Sprintf("\"%s\"", value) {
			found = true
			continue // Skip this record (remove it)
		}
		updatedRecords = append(updatedRecords, record)
	}

	if !found {
		return nil // Record doesn't exist, nothing to do
	}

	// Prepare update request with filtered records
	data := struct {
		CommonServiceItemID string `json:"CommonServiceItemID"`
		Settings            struct {
			DNS struct {
				ResourceRecordSets []Record `json:"ResourceRecordSets"`
			} `json:"DNS"`
		} `json:"Settings"`
	}{
		CommonServiceItemID: zoneID,
	}
	data.Settings.DNS.ResourceRecordSets = updatedRecords

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	url := fmt.Sprintf("%s/commonserviceitem/%s", c.apiEndpoint, zoneID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.token, c.secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	var response SingleResponse
	err = c.doJSONRequest(req, &response)
	if err != nil {
		return fmt.Errorf("failed to delete TXT record: %w", err)
	}

	if !response.IsOK {
		return fmt.Errorf("API request failed")
	}

	return nil
}

// isDomainOrSubdomain checks if the domain is a subdomain of the zone.
func isDomainOrSubdomain(domain, zone string) bool {
	return domain == zone || (len(domain) > len(zone) && domain[len(domain)-len(zone)-1:] == "."+zone)
}

// doJSONRequest makes an HTTP request and unmarshals the response.
func (c *Client) doJSONRequest(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, body)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}
