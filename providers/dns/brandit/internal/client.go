package internal

import (
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

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://portal.brandit.com/api/v3/"

// Client a BrandIT DNS API client.
type Client struct {
	apiUsername string
	apiKey      string

	baseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiUsername, apiKey string) (*Client, error) {
	if apiKey == "" || apiUsername == "" {
		return nil, errors.New("credentials missing")
	}

	return &Client{
		apiUsername: apiUsername,
		apiKey:      apiKey,
		baseURL:     defaultBaseURL,
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// ListRecords lists all records.
// https://portal.brandit.com/apidocv3#listDNSRR
func (c *Client) ListRecords(ctx context.Context, account, dnsZone string) (*ListRecordsResponse, error) {
	query := url.Values{}
	query.Add("command", "listDNSRR")
	query.Add("account", account)
	query.Add("dnszone", dnsZone)

	result := &Response[*ListRecordsResponse]{}

	err := c.do(ctx, query, result)
	if err != nil {
		return nil, err
	}

	for len(result.Response.RR) < result.Response.Total[0] {
		query.Add("first", strconv.Itoa(result.Response.Last[0]+1))

		tmp := &Response[*ListRecordsResponse]{}

		err := c.do(ctx, query, tmp)
		if err != nil {
			return nil, err
		}

		result.Response.RR = append(result.Response.RR, tmp.Response.RR...)
		result.Response.Last = tmp.Response.Last
	}

	return result.Response, nil
}

// AddRecord adds a DNS record.
// https://portal.brandit.com/apidocv3#addDNSRR
func (c *Client) AddRecord(ctx context.Context, domainName, account, newRecordID string, record Record) (*AddRecord, error) {
	value := strings.Join([]string{record.Name, strconv.Itoa(record.TTL), "IN", record.Type, record.Content}, " ")

	query := url.Values{}
	query.Add("command", "addDNSRR")
	query.Add("account", account)
	query.Add("dnszone", domainName)
	query.Add("rrdata", value)
	query.Add("key", newRecordID)

	result := &AddRecord{}

	err := c.do(ctx, query, result)
	if err != nil {
		return nil, err
	}

	result.Record = value

	return result, nil
}

// DeleteRecord deletes a DNS record.
// https://portal.brandit.com/apidocv3#deleteDNSRR
func (c *Client) DeleteRecord(ctx context.Context, domainName, account, dnsRecord, recordID string) error {
	query := url.Values{}
	query.Add("command", "deleteDNSRR")
	query.Add("account", account)
	query.Add("dnszone", domainName)
	query.Add("rrdata", dnsRecord)
	query.Add("key", recordID)

	return c.do(ctx, query, nil)
}

// StatusDomain returns the status of a domain and account associated with it.
// https://portal.brandit.com/apidocv3#statusDomain
func (c *Client) StatusDomain(ctx context.Context, domain string) (*StatusResponse, error) {
	query := url.Values{}

	query.Add("command", "statusDomain")
	query.Add("domain", domain)

	result := &Response[*StatusResponse]{}

	err := c.do(ctx, query, result)
	if err != nil {
		return nil, err
	}

	return result.Response, nil
}

func (c *Client) do(ctx context.Context, query url.Values, result any) error {
	values, err := sign(c.apiUsername, c.apiKey, query)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	//  Unmarshal the error response, because the API returns a 200 OK even if there is an error.
	var apiError APIError

	err = json.Unmarshal(raw, &apiError)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if apiError.Code > 299 || apiError.Status != "success" {
		return apiError
	}

	if result == nil {
		return nil
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func sign(apiUsername, apiKey string, query url.Values) (url.Values, error) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	canonicalRequest := fmt.Sprintf("%s%s%s", apiUsername, timestamp, defaultBaseURL)

	mac := hmac.New(sha256.New, []byte(apiKey))

	_, err := mac.Write([]byte(canonicalRequest))
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
