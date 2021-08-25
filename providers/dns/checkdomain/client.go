package checkdomain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	ns1 = "ns.checkdomain.de"
	ns2 = "ns2.checkdomain.de"
)

const domainNotFound = -1

// max page limit that the checkdomain api allows.
const maxLimit = 100

// max integer value.
const maxInt = int((^uint(0)) >> 1)

type (
	// Some fields have been omitted from the structs
	// because they are not required for this application.

	DomainListingResponse struct {
		Page     int                `json:"page"`
		Limit    int                `json:"limit"`
		Pages    int                `json:"pages"`
		Total    int                `json:"total"`
		Embedded EmbeddedDomainList `json:"_embedded"`
	}

	EmbeddedDomainList struct {
		Domains []*Domain `json:"domains"`
	}

	Domain struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	DomainResponse struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Created string `json:"created"`
		PaidUp  string `json:"payed_up"`
		Active  bool   `json:"active"`
	}

	NameserverResponse struct {
		General     NameserverGeneral `json:"general"`
		Nameservers []*Nameserver     `json:"nameservers"`
		SOA         NameserverSOA     `json:"soa"`
	}

	NameserverGeneral struct {
		IPv4       string `json:"ip_v4"`
		IPv6       string `json:"ip_v6"`
		IncludeWWW bool   `json:"include_www"`
	}

	NameserverSOA struct {
		Mail    string `json:"mail"`
		Refresh int    `json:"refresh"`
		Retry   int    `json:"retry"`
		Expiry  int    `json:"expiry"`
		TTL     int    `json:"ttl"`
	}

	Nameserver struct {
		Name string `json:"name"`
	}

	RecordListingResponse struct {
		Page     int                `json:"page"`
		Limit    int                `json:"limit"`
		Pages    int                `json:"pages"`
		Total    int                `json:"total"`
		Embedded EmbeddedRecordList `json:"_embedded"`
	}

	EmbeddedRecordList struct {
		Records []*Record `json:"records"`
	}

	Record struct {
		Name     string `json:"name"`
		Value    string `json:"value"`
		TTL      int    `json:"ttl"`
		Priority int    `json:"priority"`
		Type     string `json:"type"`
	}
)

func (d *DNSProvider) getDomainIDByName(name string) (int, error) {
	// Load from cache if exists
	d.domainIDMu.Lock()
	id, ok := d.domainIDMapping[name]
	d.domainIDMu.Unlock()
	if ok {
		return id, nil
	}

	// Find out by querying API
	domains, err := d.listDomains()
	if err != nil {
		return domainNotFound, err
	}

	// Linear search over all registered domains
	for _, domain := range domains {
		if domain.Name == name || strings.HasSuffix(name, "."+domain.Name) {
			d.domainIDMu.Lock()
			d.domainIDMapping[name] = domain.ID
			d.domainIDMu.Unlock()

			return domain.ID, nil
		}
	}

	return domainNotFound, errors.New("domain not found")
}

func (d *DNSProvider) listDomains() ([]*Domain, error) {
	req, err := d.makeRequest(http.MethodGet, "/v1/domains", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Checkdomain also provides a query param 'query' which allows filtering domains for a string.
	// But that functionality is kinda broken,
	// so we scan through the whole list of registered domains to later find the one that is of interest to us.
	q := req.URL.Query()
	q.Set("limit", strconv.Itoa(maxLimit))

	currentPage := 1
	totalPages := maxInt

	var domainList []*Domain
	for currentPage <= totalPages {
		q.Set("page", strconv.Itoa(currentPage))
		req.URL.RawQuery = q.Encode()

		var res DomainListingResponse
		if err := d.sendRequest(req, &res); err != nil {
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

func (d *DNSProvider) getNameserverInfo(domainID int) (*NameserverResponse, error) {
	req, err := d.makeRequest(http.MethodGet, fmt.Sprintf("/v1/domains/%d/nameservers", domainID), http.NoBody)
	if err != nil {
		return nil, err
	}

	res := &NameserverResponse{}
	if err := d.sendRequest(req, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (d *DNSProvider) checkNameservers(domainID int) error {
	info, err := d.getNameserverInfo(domainID)
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

func (d *DNSProvider) createRecord(domainID int, record *Record) error {
	bs, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encoding record failed: %w", err)
	}

	req, err := d.makeRequest(http.MethodPost, fmt.Sprintf("/v1/domains/%d/nameservers/records", domainID), bytes.NewReader(bs))
	if err != nil {
		return err
	}

	return d.sendRequest(req, nil)
}

// Checkdomain doesn't seem provide a way to delete records but one can replace all records at once.
// The current solution is to fetch all records and then use that list minus the record deleted as the new record list.
// TODO: Simplify this function once Checkdomain do provide the functionality.
func (d *DNSProvider) deleteTXTRecord(domainID int, recordName, recordValue string) error {
	domainInfo, err := d.getDomainInfo(domainID)
	if err != nil {
		return err
	}

	nsInfo, err := d.getNameserverInfo(domainID)
	if err != nil {
		return err
	}

	allRecords, err := d.listRecords(domainID, "")
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

	return d.replaceRecords(domainID, recordsToKeep)
}

func (d *DNSProvider) getDomainInfo(domainID int) (*DomainResponse, error) {
	req, err := d.makeRequest(http.MethodGet, fmt.Sprintf("/v1/domains/%d", domainID), http.NoBody)
	if err != nil {
		return nil, err
	}

	var res DomainResponse
	err = d.sendRequest(req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (d *DNSProvider) listRecords(domainID int, recordType string) ([]*Record, error) {
	req, err := d.makeRequest(http.MethodGet, fmt.Sprintf("/v1/domains/%d/nameservers/records", domainID), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", strconv.Itoa(maxLimit))
	if recordType != "" {
		q.Set("type", recordType)
	}

	currentPage := 1
	totalPages := maxInt

	var recordList []*Record
	for currentPage <= totalPages {
		q.Set("page", strconv.Itoa(currentPage))
		req.URL.RawQuery = q.Encode()

		var res RecordListingResponse
		if err := d.sendRequest(req, &res); err != nil {
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

func (d *DNSProvider) replaceRecords(domainID int, records []*Record) error {
	bs, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("encoding record failed: %w", err)
	}

	req, err := d.makeRequest(http.MethodPut, fmt.Sprintf("/v1/domains/%d/nameservers/records", domainID), bytes.NewReader(bs))
	if err != nil {
		return err
	}

	return d.sendRequest(req, nil)
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

func (d *DNSProvider) makeRequest(method, resource string, body io.Reader) (*http.Request, error) {
	uri, err := d.config.Endpoint.Parse(resource)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, uri.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.config.Token)
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (d *DNSProvider) sendRequest(req *http.Request, result interface{}) error {
	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if err = checkResponse(resp); err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return fmt.Errorf("unmarshaling %T error [status code=%d]: %w: %s", result, resp.StatusCode, err, string(raw))
	}
	return nil
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	if resp.Body == nil {
		return fmt.Errorf("response body is nil, status code=%d", resp.StatusCode)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read body: status code=%d, error=%w", resp.StatusCode, err)
	}

	return fmt.Errorf("status code=%d: %s", resp.StatusCode, string(raw))
}
