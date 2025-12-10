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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"golang.org/x/oauth2"
)

const (
	ns1 = "ns.checkdomain.de"
	ns2 = "ns2.checkdomain.de"
)

// DefaultEndpoint the default API endpoint.
const DefaultEndpoint = "https://api.checkdomain.de"

const domainNotFound = -1

// max page limit that the checkdomain api allows.
const maxLimit = 100

// max integer value.
const maxInt = int((^uint(0)) >> 1)

// Client the Autodns API client.
type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client

	domainIDMapping map[string]int
	domainIDMu      sync.Mutex
}

// NewClient creates a new Client.
func NewClient(hc *http.Client) *Client {
	baseURL, _ := url.Parse(DefaultEndpoint)

	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{
		BaseURL:         baseURL,
		httpClient:      hc,
		domainIDMapping: make(map[string]int),
	}
}

func (c *Client) GetDomainIDByName(ctx context.Context, name string) (int, error) {
	// Load from cache if exists
	c.domainIDMu.Lock()
	id, ok := c.domainIDMapping[name]
	c.domainIDMu.Unlock()

	if ok {
		return id, nil
	}

	// Find out by querying API
	domains, err := c.listDomains(ctx)
	if err != nil {
		return domainNotFound, err
	}

	// Linear search over all registered domains
	for _, domain := range domains {
		if domain.Name == name || strings.HasSuffix(name, "."+domain.Name) {
			c.domainIDMu.Lock()
			c.domainIDMapping[name] = domain.ID
			c.domainIDMu.Unlock()

			return domain.ID, nil
		}
	}

	return domainNotFound, errors.New("domain not found")
}

func (c *Client) listDomains(ctx context.Context) ([]*Domain, error) {
	endpoint := c.BaseURL.JoinPath("v1", "domains")

	// Checkdomain also provides a query param 'query' which allows filtering domains for a string.
	// But that functionality is kinda broken,
	// so we scan through the whole list of registered domains to later find the one that is of interest to us.
	q := endpoint.Query()
	q.Set("limit", strconv.Itoa(maxLimit))

	currentPage := 1
	totalPages := maxInt

	var domainList []*Domain

	for currentPage <= totalPages {
		q.Set("page", strconv.Itoa(currentPage))
		endpoint.RawQuery = q.Encode()

		req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}

		var res DomainListingResponse
		if err := c.do(req, &res); err != nil {
			return nil, fmt.Errorf("failed to send domain listing request: %w", err)
		}

		// This is the first response,
		// so we update totalPages and allocate the slice memory.
		if totalPages == maxInt {
			totalPages = res.Pages
			domainList = make([]*Domain, 0, res.Total)
		}

		domainList = append(domainList, res.Embedded.Domains...)
		currentPage++
	}

	return domainList, nil
}

func (c *Client) getNameserverInfo(ctx context.Context, domainID int) (*NameserverResponse, error) {
	endpoint := c.BaseURL.JoinPath("v1", "domains", strconv.Itoa(domainID), "nameservers")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	res := &NameserverResponse{}
	if err := c.do(req, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) CheckNameservers(ctx context.Context, domainID int) error {
	info, err := c.getNameserverInfo(ctx, domainID)
	if err != nil {
		return err
	}

	var found1, found2 bool

	for _, item := range info.Nameservers {
		switch item.Name {
		case ns1:
			found1 = true
		case ns2:
			found2 = true
		}
	}

	if !found1 || !found2 {
		return errors.New("not using checkdomain nameservers, can not update records")
	}

	return nil
}

func (c *Client) CreateRecord(ctx context.Context, domainID int, record *Record) error {
	endpoint := c.BaseURL.JoinPath("v1", "domains", strconv.Itoa(domainID), "nameservers", "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// DeleteTXTRecord Checkdomain doesn't seem provide a way to delete records but one can replace all records at once.
// The current solution is to fetch all records and then use that list minus the record deleted as the new record list.
// TODO: Simplify this function once Checkdomain do provide the functionality.
func (c *Client) DeleteTXTRecord(ctx context.Context, domainID int, recordName, recordValue string) error {
	domainInfo, err := c.getDomainInfo(ctx, domainID)
	if err != nil {
		return err
	}

	nsInfo, err := c.getNameserverInfo(ctx, domainID)
	if err != nil {
		return err
	}

	allRecords, err := c.listRecords(ctx, domainID, "")
	if err != nil {
		return err
	}

	recordName = strings.TrimSuffix(recordName, "."+domainInfo.Name+".")

	var recordsToKeep []*Record

	// Find and delete matching records
	for _, record := range allRecords {
		if skipRecord(recordName, recordValue, record, nsInfo) {
			continue
		}

		// Checkdomain API can return records without any TTL set (indicated by the value of 0).
		// The API Call to replace the records would fail if we wouldn't specify a value.
		// Thus, we use the default TTL queried beforehand
		if record.TTL == 0 {
			record.TTL = nsInfo.SOA.TTL
		}

		recordsToKeep = append(recordsToKeep, record)
	}

	return c.replaceRecords(ctx, domainID, recordsToKeep)
}

func (c *Client) getDomainInfo(ctx context.Context, domainID int) (*DomainResponse, error) {
	endpoint := c.BaseURL.JoinPath("v1", "domains", strconv.Itoa(domainID))

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var res DomainResponse

	err = c.do(req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) listRecords(ctx context.Context, domainID int, recordType string) ([]*Record, error) {
	endpoint := c.BaseURL.JoinPath("v1", "domains", strconv.Itoa(domainID), "nameservers", "records")

	q := endpoint.Query()
	q.Set("limit", strconv.Itoa(maxLimit))

	if recordType != "" {
		q.Set("type", recordType)
	}

	currentPage := 1
	totalPages := maxInt

	var recordList []*Record

	for currentPage <= totalPages {
		q.Set("page", strconv.Itoa(currentPage))
		endpoint.RawQuery = q.Encode()

		req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		var res RecordListingResponse
		if err := c.do(req, &res); err != nil {
			return nil, fmt.Errorf("failed to send record listing request: %w", err)
		}

		// This is the first response, so we update totalPages and allocate the slice memory.
		if totalPages == maxInt {
			totalPages = res.Pages
			recordList = make([]*Record, 0, res.Total)
		}

		recordList = append(recordList, res.Embedded.Records...)
		currentPage++
	}

	return recordList, nil
}

func (c *Client) replaceRecords(ctx context.Context, domainID int, records []*Record) error {
	endpoint := c.BaseURL.JoinPath("v1", "domains", strconv.Itoa(domainID), "nameservers", "records")

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, records)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func (c *Client) CleanCache(fqdn string) {
	c.domainIDMu.Lock()
	delete(c.domainIDMapping, fqdn)
	c.domainIDMu.Unlock()
}

func skipRecord(recordName, recordValue string, record *Record, nsInfo *NameserverResponse) bool {
	// Skip empty records
	if record.Value == "" {
		return true
	}

	// Skip some special records, otherwise we would get a "Nameserver update failed"
	if record.Type == "SOA" || record.Type == "NS" || record.Name == "@" || (nsInfo.General.IncludeWWW && record.Name == "www") {
		return true
	}

	nameMatch := recordName == "" || record.Name == recordName
	valueMatch := recordValue == "" || record.Value == recordValue

	// Skip our matching record
	if record.Type == "TXT" && nameMatch && valueMatch {
		return true
	}

	return false
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

func OAuthStaticAccessToken(client *http.Client, accessToken string) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	client.Transport = &oauth2.Transport{
		Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
		Base:   client.Transport,
	}

	return client
}
