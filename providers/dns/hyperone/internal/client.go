package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://api.hyperone.com/v2"

const defaultLocationID = "pl-waw-1"

type signer interface {
	GetJWT() (string, error)
}

// Client the HyperOne client.
type Client struct {
	passport *Passport
	signer   signer

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new HyperOne client.
func NewClient(apiEndpoint, locationID string, passport *Passport) (*Client, error) {
	if passport == nil {
		return nil, errors.New("the passport is missing")
	}

	projectID, err := passport.ExtractProjectID()
	if err != nil {
		return nil, err
	}

	if apiEndpoint == "" {
		apiEndpoint = defaultBaseURL
	}

	baseURL, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, err
	}

	tokenSigner := &TokenSigner{
		PrivateKey: passport.PrivateKey,
		KeyID:      passport.CertificateID,
		Audience:   apiEndpoint,
		Issuer:     passport.Issuer,
		Subject:    passport.SubjectID,
	}

	if locationID == "" {
		locationID = defaultLocationID
	}

	client := &Client{
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    baseURL.JoinPath("dns", locationID, "project", projectID),
		passport:   passport,
		signer:     tokenSigner,
	}

	return client, nil
}

// FindRecordset looks for recordset with given recordType and name and returns it.
// In case if recordset is not found returns nil.
// https://api.hyperone.com/v2/docs#operation/dns_project_zone_recordset_list
func (c *Client) FindRecordset(ctx context.Context, zoneID, recordType, name string) (*Recordset, error) {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone/{zoneId}/recordset
	endpoint := c.baseURL.JoinPath("zone", zoneID, "recordset")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var recordSets []Recordset

	err = c.do(req, &recordSets)
	if err != nil {
		return nil, fmt.Errorf("failed to get recordsets from server: %w", err)
	}

	for _, v := range recordSets {
		if v.RecordType == recordType && v.Name == name {
			return &v, nil
		}
	}

	// when recordset is not present returns nil, but error is not thrown
	return nil, nil
}

// CreateRecordset creates recordset and record with given value within one request.
// https://api.hyperone.com/v2/docs#operation/dns_project_zone_recordset_create
func (c *Client) CreateRecordset(ctx context.Context, zoneID, recordType, name, recordValue string, ttl int) (*Recordset, error) {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone/{zoneId}/recordset
	endpoint := c.baseURL.JoinPath("zone", zoneID, "recordset")

	recordsetInput := Recordset{
		RecordType: recordType,
		Name:       name,
		TTL:        ttl,
		Record:     &Record{Content: recordValue},
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, recordsetInput)
	if err != nil {
		return nil, err
	}

	var recordsetResponse Recordset

	err = c.do(req, &recordsetResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create recordset: %w", err)
	}

	return &recordsetResponse, nil
}

// DeleteRecordset deletes a recordset.
// https://api.hyperone.com/v2/docs#operation/dns_project_zone_recordset_delete
func (c *Client) DeleteRecordset(ctx context.Context, zoneID, recordsetID string) error {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone/{zoneId}/recordset/{recordsetId}
	endpoint := c.baseURL.JoinPath("zone", zoneID, "recordset", recordsetID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// GetRecords gets all records within specified recordset.
// https://api.hyperone.com/v2/docs#operation/dns_project_zone_recordset_record_list
func (c *Client) GetRecords(ctx context.Context, zoneID, recordsetID string) ([]Record, error) {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone/{zoneId}/recordset/{recordsetId}/record
	endpoint := c.baseURL.JoinPath("zone", zoneID, "recordset", recordsetID, "record")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var records []Record

	err = c.do(req, &records)
	if err != nil {
		return nil, fmt.Errorf("failed to get records from server: %w", err)
	}

	return records, err
}

// CreateRecord creates a record.
// https://api.hyperone.com/v2/docs#operation/dns_project_zone_recordset_record_create
func (c *Client) CreateRecord(ctx context.Context, zoneID, recordsetID, recordContent string) (*Record, error) {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone/{zoneId}/recordset/{recordsetId}/record
	endpoint := c.baseURL.JoinPath("zone", zoneID, "recordset", recordsetID, "record")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, Record{Content: recordContent})
	if err != nil {
		return nil, err
	}

	var recordResponse Record

	err = c.do(req, &recordResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to set record: %w", err)
	}

	return &recordResponse, nil
}

// DeleteRecord deletes a record.
// https://api.hyperone.com/v2/docs#operation/dns_project_zone_recordset_record_delete
func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordsetID, recordID string) error {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone/{zoneId}/recordset/{recordsetId}/record/{recordId}
	endpoint := c.baseURL.JoinPath("zone", zoneID, "recordset", recordsetID, "record", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// FindZone looks for DNS Zone and returns nil if it does not exist.
func (c *Client) FindZone(ctx context.Context, name string) (*Zone, error) {
	zones, err := c.GetZones(ctx)
	if err != nil {
		return nil, err
	}

	for _, zone := range zones {
		if zone.DNSName == name {
			return &zone, nil
		}
	}

	return nil, fmt.Errorf("failed to find zone for %s", name)
}

// GetZones gets all user's zones.
// https://api.hyperone.com/v2/docs#operation/dns_project_zone_list
func (c *Client) GetZones(ctx context.Context) ([]Zone, error) {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone
	endpoint := c.baseURL.JoinPath("zone")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var zones []Zone

	err = c.do(req, &zones)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available zones: %w", err)
	}

	return zones, nil
}

func (c *Client) do(req *http.Request, result any) error {
	jwt, err := c.signer.GetJWT()
	if err != nil {
		return fmt.Errorf("failed to sign the request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return parseError(req, resp)
	}

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if err = json.Unmarshal(raw, result); err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func newJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func parseError(req *http.Request, resp *http.Response) error {
	var msg string
	if resp.StatusCode == http.StatusForbidden {
		msg = "forbidden: check if service account you are trying to use has permissions required for managing DNS"
	} else {
		msg = "unknown error"
	}

	return fmt.Errorf("%s: %w", msg, errutils.NewUnexpectedResponseStatusCodeError(req, resp))
}
