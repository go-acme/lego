package internal

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://portal.brandit.com/api/v3/"

// StatusSuccess expected status text when success.
const StatusSuccess = "success"

// Client a BrandIT DNS API client.
type Client struct {
	apiUsername string
	apiKey      string
	BaseURL     string
	HTTPClient  *http.Client
}

// NewClient creates a new Client.
func NewClient(apiUsername, apiKey string) (*Client, error) {
	if apiKey == "" || apiUsername == "" {
		return nil, errors.New("credentials missing")
	}

	return &Client{
		apiUsername: apiUsername,
		apiKey:      apiKey,
		BaseURL:     defaultBaseURL,
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// ListRecords lists all records.
// https://portal.brandit.com/apidocv3#listDNSRR
func (c *Client) ListRecords(account, dnsZone string) (*ListRecords, error) {
	// Create a new query
	query := url.Values{}
	query.Add("command", "listDNSRR")
	query.Add("account", account)
	query.Add("dnszone", dnsZone)

	result := &ListRecords{}

	err := c.do(query, result)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	for len(result.Response.RR) < result.Response.Total[0] {
		query.Add("first", fmt.Sprint(result.Response.Last[0]+1))

		tmp := &ListRecords{}
		err := c.do(query, tmp)
		if err != nil {
			return nil, fmt.Errorf("do: %w", err)
		}

		result.Response.RR = append(result.Response.RR, tmp.Response.RR...)
		result.Response.Last = tmp.Response.Last
	}

	return result, nil
}

// AddRecord adds a DNS record.
// https://portal.brandit.com/apidocv3#addDNSRR
func (c *Client) AddRecord(domainName, account, newRecordID string, record Record) (*AddRecord, error) {
	// Create a new query

	query := url.Values{}
	query.Add("command", "addDNSRR")
	query.Add("account", account)
	query.Add("dnszone", domainName)
	query.Add("rrdata", strings.Join([]string{record.Name, fmt.Sprint(record.TTL), "IN", record.Type, record.Content}, " "))
	query.Add("key", newRecordID)

	result := &AddRecord{}

	err := c.do(query, result)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}
	result.Record = strings.Join([]string{record.Name, fmt.Sprint(record.TTL), "IN", record.Type, record.Content}, " ")

	return result, nil
}

func (c *Client) PresentRecord(domainName string, record Record) (*AddRecord, error) {
	// find the account associated with the domain
	account, err := c.StatusDomain(domainName)
	if err != nil {
		return nil, fmt.Errorf("status domain: %w", err)
	}

	// Find the next record id
	recordID, err := c.ListRecords(account.Response.Registrar[0], domainName)
	if err != nil {
		return nil, fmt.Errorf("list records: %w", err)
	}

	result, err := c.AddRecord(domainName, account.Response.Registrar[0], fmt.Sprint(recordID.Response.Total[0]), record)
	if err != nil {
		return nil, fmt.Errorf("add record: %w", err)
	}

	return result, nil
}

func (c *Client) CleanUpRecord(domainName, dnsRecord string) (*DeleteRecord, error) {
	// find the account associated with the domain
	account, err := c.StatusDomain(domainName)
	if err != nil {
		return nil, fmt.Errorf("status domain: %w", err)
	}

	records, err := c.ListRecords(account.Response.Registrar[0], domainName)
	if err != nil {
		return nil, fmt.Errorf("list records: %w", err)
	}
	var recordID int
	for i, r := range records.Response.Rr {
		if r == dnsRecord {
			recordID = i
		}
	}
	result, err := c.DeleteRecord(domainName, account.Response.Registrar[0], dnsRecord, fmt.Sprint(recordID))
	if err != nil {
		return nil, fmt.Errorf("delete record: %w", err)
	}

	return result, nil
}

// DeleteRecord deletes a DNS record.
// https://portal.brandit.com/apidocv3#deleteDNSRR
func (c *Client) DeleteRecord(domainName, account, dnsRecord, recordID string) (*DeleteRecord, error) {
	// Create a new query
	query := url.Values{}
	query.Add("command", "deleteDNSRR")
	query.Add("account", account)
	query.Add("dnszone", domainName)
	query.Add("rrdata", dnsRecord)
	query.Add("key", recordID)

	result := &DeleteRecord{}

	err := c.do(query, result)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	return result, nil
}

// StatusDomain returns the status of a domain and account associated with it.
// https://portal.brandit.com/apidocv3#statusDomain
func (c *Client) StatusDomain(domain string) (*StatusDomain, error) {
	// Create a new query
	query := url.Values{}

	query.Add("command", "statusDomain")
	query.Add("domain", domain)

	result := &StatusDomain{}

	err := c.do(query, result)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	return result, nil
}

func (c *Client) do(query url.Values, result any) error {
	// Add signature
	v, err := sign(c.apiUsername, c.apiKey, query)
	if err != nil {
		return fmt.Errorf("signature: %w", err)
	}

	resp, err := c.HTTPClient.PostForm(c.BaseURL, v)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	//  Unmarshal the error response, because the API returns a 200 OK even if there is an error.
	var Err Errors
	err = json.Unmarshal(all, &Err)
	if err != nil {
		return fmt.Errorf("unmarshal response body: %w", err)
	}
	if Err.Code > 299 {
		return fmt.Errorf("response code: %d, message: %s", Err.Code, Err.Error)
	}

	err = json.Unmarshal(all, result)
	if err != nil {
		return fmt.Errorf("unmarshal response body: %w", err)
	}

	return nil
}

func sign(apiUsername, apiKey string, query url.Values) (url.Values, error) {
	location, err := time.LoadLocation("GMT")
	if err != nil {
		return nil, fmt.Errorf("time location: %w", err)
	}

	timestamp := time.Now().In(location).Format("2006-01-02T15:04:05Z")

	canonicalRequest := fmt.Sprintf("%s%s%s", apiUsername, timestamp, defaultBaseURL)

	mac := hmac.New(sha256.New, []byte(apiKey))
	_, err = mac.Write([]byte(canonicalRequest))
	if err != nil {
		return nil, err
	}

	hashed := mac.Sum(nil)
	signature := hex.EncodeToString(hashed)

	query.Add("user", apiUsername)
	query.Add("timestamp", timestamp)
	query.Add("signature", signature)

	return query, nil
}
