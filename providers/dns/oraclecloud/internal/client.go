package internal

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client Oracle Cloud DNS API client.
type Client struct {
	CompartmentID string
	httpClient    *http.Client
	privateKey    *rsa.PrivateKey
	keyID         string
	baseURL       string
}

// Record represents a DNS record.
type Record struct {
	Domain       string `json:"domain,omitempty"`
	RecordHash   string `json:"recordHash,omitempty"`
	IsProtected  bool   `json:"isProtected,omitempty"`
	RData        string `json:"rdata,omitempty"`
	RRSetVersion string `json:"rrsetVersion,omitempty"`
	RType        string `json:"rtype,omitempty"`
	TTL          int    `json:"ttl,omitempty"`
}

// RecordOperation represents a DNS record operation.
type RecordOperation struct {
	Operation   string `json:"operation,omitempty"`
	Domain      string `json:"domain,omitempty"`
	RecordHash  string `json:"recordHash,omitempty"`
	RData       string `json:"rdata,omitempty"`
	RType       string `json:"rtype,omitempty"`
	TTL         int    `json:"ttl,omitempty"`
	IsProtected bool   `json:"isProtected,omitempty"`
}

// PatchRecordsRequest represents a request to patch DNS records.
type PatchRecordsRequest struct {
	Items []RecordOperation `json:"items"`
}

// RecordsResponse represents a response containing DNS records.
type RecordsResponse struct {
	Items []Record `json:"items"`
}

// NewClient creates a new Oracle Cloud DNS client.
func NewClient(httpClient *http.Client, privateKey *rsa.PrivateKey, keyID string, region string, compartmentID string) *Client {
	baseURL := fmt.Sprintf("https://dns.%s.oraclecloud.com/20180115", region)

	return &Client{
		CompartmentID: compartmentID,
		httpClient:    httpClient,
		privateKey:    privateKey,
		keyID:         keyID,
		baseURL:       baseURL,
	}
}

// GetDomainRecords gets all records for a domain.
func (c *Client) GetDomainRecords(ctx context.Context, zoneNameOrID, domain, recordType string) (*RecordsResponse, error) {
	endpoint := fmt.Sprintf("%s/zones/%s/records/%s", c.baseURL, zoneNameOrID, domain)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("compartmentId", c.CompartmentID)
	if recordType != "" {
		q.Add("rtype", recordType)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")

	if err := c.signRequest(req); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d, %s", resp.StatusCode, string(body))
	}

	var records RecordsResponse
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &records, nil
}

// PatchDomainRecords updates records for a domain.
func (c *Client) PatchDomainRecords(ctx context.Context, zoneNameOrID, domain string, records PatchRecordsRequest) error {
	endpoint := fmt.Sprintf("%s/zones/%s/records/%s", c.baseURL, zoneNameOrID, domain)

	body, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("compartmentId", c.CompartmentID)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")

	if err := c.signRequest(req); err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %d, %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListDomainRecords is a helper method to list all records for a domain.
func (c *Client) ListDomainRecords(ctx context.Context, zoneNameOrID string) (*RecordsResponse, error) {
	endpoint := fmt.Sprintf("%s/zones/%s/records", c.baseURL, zoneNameOrID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("compartmentId", c.CompartmentID)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")

	if err := c.signRequest(req); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d, %s", resp.StatusCode, string(body))
	}

	var records RecordsResponse
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &records, nil
}

// DeleteDomainRecord deletes a specific record by its hash.
func (c *Client) DeleteDomainRecord(ctx context.Context, zoneNameOrID, domain string, recordHash string) error {
	recordOperation := RecordOperation{
		RecordHash: recordHash,
		Operation:  "REMOVE",
	}

	request := PatchRecordsRequest{
		Items: []RecordOperation{recordOperation},
	}

	return c.PatchDomainRecords(ctx, zoneNameOrID, domain, request)
}

// FindAndDeleteDomainRecord finds and deletes TXT records matching the given value.
func (c *Client) FindAndDeleteDomainRecord(ctx context.Context, zoneNameOrID, domain, expectedValue string) error {
	// Get existing TXT records for the domain
	domainRecords, err := c.GetDomainRecords(ctx, zoneNameOrID, domain, "TXT")
	if err != nil {
		return fmt.Errorf("error getting domain records: %w", err)
	}

	if len(domainRecords.Items) == 0 {
		// No records found, nothing to delete
		return nil
	}

	var recordsToDelete []string
	// Look for records matching our value
	for _, record := range domainRecords.Items {
		// Strip quotes for comparison
		rdata := strings.ReplaceAll(record.RData, `\"`, "")
		rdata = strings.ReplaceAll(rdata, `"`, "")

		// Check for exact match or if it contains our value
		if rdata == expectedValue || strings.Contains(rdata, expectedValue) {
			recordsToDelete = append(recordsToDelete, record.RecordHash)
		}
	}

	if len(recordsToDelete) == 0 {
		return nil
	}

	// Delete all matching records
	var operations []RecordOperation
	for _, hash := range recordsToDelete {
		operations = append(operations, RecordOperation{
			RecordHash: hash,
			Operation:  "REMOVE",
		})
	}

	request := PatchRecordsRequest{
		Items: operations,
	}

	err = c.PatchDomainRecords(ctx, zoneNameOrID, domain, request)
	if err != nil {
		return fmt.Errorf("error deleting domain records: %w", err)
	}

	return nil
}

// signRequest signs the request according to OCI signing specifications.
func (c *Client) signRequest(req *http.Request) error {
	// Required headers for OCI API signature
	date := time.Now().UTC().Format(http.TimeFormat)
	req.Header.Set("Date", date)

	// Set host header if not already set
	if req.Host == "" && req.URL != nil {
		req.Host = req.URL.Host
	}

	// Set the host header explicitly
	req.Header.Set("host", req.Host)

	// Content hash (required for PUT, POST, PATCH requests with body)
	if req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch {
		if req.Body != nil {
			var bodyBytes []byte
			// Save the body
			if req.GetBody != nil {
				var err error
				bodyReadCloser, err := req.GetBody()
				if err != nil {
					return err
				}
				bodyBytes, err = io.ReadAll(bodyReadCloser)
				bodyReadCloser.Close()
				if err != nil {
					return err
				}
			} else if req.Body != nil {
				var err error
				bodyBytes, err = io.ReadAll(req.Body)
				if err != nil {
					return err
				}
				// Replace the body
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				// Set GetBody for reuse
				req.GetBody = func() (io.ReadCloser, error) {
					return io.NopCloser(bytes.NewReader(bodyBytes)), nil
				}
			}

			// Calculate SHA256 of body
			bodySha256 := sha256.Sum256(bodyBytes)
			sha256Base64 := base64.StdEncoding.EncodeToString(bodySha256[:])
			req.Header.Set("x-content-sha256", sha256Base64)
		} else {
			// Empty body
			bodySha256 := sha256.Sum256([]byte{})
			sha256Base64 := base64.StdEncoding.EncodeToString(bodySha256[:])
			req.Header.Set("x-content-sha256", sha256Base64)
		}

		// Make sure content-length is set properly
		if req.ContentLength >= 0 {
			req.Header.Set("content-length", fmt.Sprintf("%d", req.ContentLength))
		} else {
			// If content length is not set but there's a body, attempt to set it
			if req.Body != nil {
				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					return err
				}
				contentLength := len(bodyBytes)
				req.ContentLength = int64(contentLength)
				req.Header.Set("content-length", fmt.Sprintf("%d", contentLength))
				// Reset the body
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				req.GetBody = func() (io.ReadCloser, error) {
					return io.NopCloser(bytes.NewReader(bodyBytes)), nil
				}
			} else {
				// Empty body
				req.ContentLength = 0
				req.Header.Set("content-length", "0")
			}
		}
	}

	// Generate the signature
	signingString, headers, err := c.getSigningString(req)
	if err != nil {
		return err
	}

	hasher := sha256.New()
	hasher.Write([]byte(signingString))
	hashed := hasher.Sum(nil)

	signature, err := rsa.SignPKCS1v15(nil, c.privateKey, crypto.SHA256, hashed)
	if err != nil {
		return err
	}

	// Add the authorization header
	signatureB64 := base64.StdEncoding.EncodeToString(signature)
	authHeader := fmt.Sprintf(`Signature version="1",keyId="%s",algorithm="rsa-sha256",headers="%s",signature="%s"`,
		c.keyID, strings.Join(headers, " "), signatureB64)

	req.Header.Set("Authorization", authHeader)

	return nil
}

// getSigningString constructs the string to be signed according to OCI specifications.
func (c *Client) getSigningString(req *http.Request) (string, []string, error) {
	var signHeaders []string
	var headerValues []string

	// Add (request-target) always
	target := fmt.Sprintf("%s %s", strings.ToLower(req.Method), req.URL.Path)
	if req.URL.RawQuery != "" {
		target = fmt.Sprintf("%s?%s", target, req.URL.RawQuery)
	}
	signHeaders = append(signHeaders, "(request-target)")
	headerValues = append(headerValues, fmt.Sprintf("(request-target): %s", target))

	// Always include host header in signing
	signHeaders = append(signHeaders, "host")
	headerValues = append(headerValues, fmt.Sprintf("host: %s", req.Host))

	// Always include date header in signing
	signHeaders = append(signHeaders, "date")
	headerValues = append(headerValues, fmt.Sprintf("date: %s", req.Header.Get("date")))

	// For requests with a body, include required headers
	if req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch {
		// Content-Type is required for requests with a body
		if contentType := req.Header.Get("content-type"); contentType != "" {
			signHeaders = append(signHeaders, "content-type")
			headerValues = append(headerValues, fmt.Sprintf("content-type: %s", contentType))
		}

		// x-content-sha256 is required for requests with a body
		if sha256 := req.Header.Get("x-content-sha256"); sha256 != "" {
			signHeaders = append(signHeaders, "x-content-sha256")
			headerValues = append(headerValues, fmt.Sprintf("x-content-sha256: %s", sha256))
		}

		// Content-Length is required for requests with a body
		if contentLength := req.Header.Get("content-length"); contentLength != "" {
			signHeaders = append(signHeaders, "content-length")
			headerValues = append(headerValues, fmt.Sprintf("content-length: %s", contentLength))
		}
	}

	signatureString := strings.Join(headerValues, "\n")

	return signatureString, signHeaders, nil
}
