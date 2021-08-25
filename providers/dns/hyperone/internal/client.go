package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

const defaultBaseURL = "https://api.hyperone.com/v2"

const defaultLocationID = "pl-waw-1"

type signer interface {
	GetJWT() (string, error)
}

// Client the HyperOne client.
type Client struct {
	HTTPClient *http.Client

	apiEndpoint string
	locationID  string
	projectID   string

	passport *Passport
	signer   signer
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

	baseURL := defaultBaseURL
	if apiEndpoint != "" {
		baseURL = apiEndpoint
	}

	tokenSigner := &TokenSigner{
		PrivateKey: passport.PrivateKey,
		KeyID:      passport.CertificateID,
		Audience:   baseURL,
		Issuer:     passport.Issuer,
		Subject:    passport.SubjectID,
	}

	client := &Client{
		HTTPClient:  &http.Client{Timeout: 5 * time.Second},
		apiEndpoint: baseURL,
		locationID:  locationID,
		passport:    passport,
		projectID:   projectID,
		signer:      tokenSigner,
	}

	if client.locationID == "" {
		client.locationID = defaultLocationID
	}

	return client, nil
}

// FindRecordset looks for recordset with given recordType and name and returns it.
// In case if recordset is not found returns nil.
// https://api.hyperone.com/v2/docs#operation/dns_project_zone_recordset_list
func (c *Client) FindRecordset(zoneID, recordType, name string) (*Recordset, error) {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone/{zoneId}/recordset
	resourceURL := path.Join("dns", c.locationID, "project", c.projectID, "zone", zoneID, "recordset")

	req, err := c.createRequest(http.MethodGet, resourceURL, nil)
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
func (c *Client) CreateRecordset(zoneID, recordType, name, recordValue string, ttl int) (*Recordset, error) {
	recordsetInput := Recordset{
		RecordType: recordType,
		Name:       name,
		TTL:        ttl,
		Record:     &Record{Content: recordValue},
	}

	requestBody, err := json.Marshal(recordsetInput)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recordset: %w", err)
	}

	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone/{zoneId}/recordset
	resourceURL := path.Join("dns", c.locationID, "project", c.projectID, "zone", zoneID, "recordset")

	req, err := c.createRequest(http.MethodPost, resourceURL, bytes.NewBuffer(requestBody))
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
func (c *Client) DeleteRecordset(zoneID string, recordsetID string) error {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone/{zoneId}/recordset/{recordsetId}
	resourceURL := path.Join("dns", c.locationID, "project", c.projectID, "zone", zoneID, "recordset", recordsetID)

	req, err := c.createRequest(http.MethodDelete, resourceURL, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// GetRecords gets all records within specified recordset.
// https://api.hyperone.com/v2/docs#operation/dns_project_zone_recordset_record_list
func (c *Client) GetRecords(zoneID string, recordsetID string) ([]Record, error) {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone/{zoneId}/recordset/{recordsetId}/record
	resourceURL := path.Join("dns", c.locationID, "project", c.projectID, "zone", zoneID, "recordset", recordsetID, "record")

	req, err := c.createRequest(http.MethodGet, resourceURL, nil)
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
func (c *Client) CreateRecord(zoneID, recordsetID, recordContent string) (*Record, error) {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone/{zoneId}/recordset/{recordsetId}/record
	resourceURL := path.Join("dns", c.locationID, "project", c.projectID, "zone", zoneID, "recordset", recordsetID, "record")

	requestBody, err := json.Marshal(Record{Content: recordContent})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %w", err)
	}

	req, err := c.createRequest(http.MethodPost, resourceURL, bytes.NewBuffer(requestBody))
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
func (c *Client) DeleteRecord(zoneID, recordsetID, recordID string) error {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone/{zoneId}/recordset/{recordsetId}/record/{recordId}
	resourceURL := path.Join("dns", c.locationID, "project", c.projectID, "zone", zoneID, "recordset", recordsetID, "record", recordID)

	req, err := c.createRequest(http.MethodDelete, resourceURL, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// FindZone looks for DNS Zone and returns nil if it does not exist.
func (c *Client) FindZone(name string) (*Zone, error) {
	zones, err := c.GetZones()
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
func (c *Client) GetZones() ([]Zone, error) {
	// https://api.hyperone.com/v2/dns/{locationId}/project/{projectId}/zone
	resourceURL := path.Join("dns", c.locationID, "project", c.projectID, "zone")

	req, err := c.createRequest(http.MethodGet, resourceURL, nil)
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

func (c *Client) createRequest(method, uri string, body io.Reader) (*http.Request, error) {
	baseURL, err := url.Parse(c.apiEndpoint)
	if err != nil {
		return nil, err
	}

	endpoint, err := baseURL.Parse(path.Join(baseURL.Path, uri))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	jwt, err := c.signer.GetJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to sign the request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	err = checkResponse(resp)
	if err != nil {
		return err
	}

	if v == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	if err = json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("unmarshaling %T error: %w: %s", v, err, string(raw))
	}

	return nil
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode/100 == 2 {
		return nil
	}

	var msg string
	if resp.StatusCode == http.StatusForbidden {
		msg = "forbidden: check if service account you are trying to use has permissions required for managing DNS"
	} else {
		msg = fmt.Sprintf("%d: unknown error", resp.StatusCode)
	}

	// add response body to error message if not empty
	responseBody, _ := io.ReadAll(resp.Body)
	if len(responseBody) > 0 {
		msg = fmt.Sprintf("%s: %s", msg, string(responseBody))
	}

	return errors.New(msg)
}
