package checkdomain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// max page limit that the checkdomain api allows
const maxLimit = 100

// max integer value
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
		PaidUp  string `json:"payed_up"` //nolint
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

func (p *DNSProvider) listDomains() ([]*Domain, error) {
	req, err := p.makeRequest(http.MethodGet, "/v1/domains", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}

	// Checkdomain also provides a query param 'query' which allows filtering
	// domains for a string. But that functionality is kinda broken, so we scan
	// through the whole list of registered domains to later find the one that is
	// of interest to us.

	q := req.URL.Query()
	q.Set("limit", strconv.Itoa(maxLimit))

	var (
		currentPage = 1
		totalPages  = maxInt
	)

	var domainList []*Domain
	for currentPage <= totalPages {
		q.Set("page", strconv.Itoa(currentPage))
		req.URL.RawQuery = q.Encode()

		var res DomainListingResponse
		if err := p.sendRequest(req, &res); err != nil {
			return nil, fmt.Errorf("failed to send domain listing request: %v", err)
		}

		// This is the first response, so we update
		// totalPages and allocate the slice memory.
		if totalPages == maxInt {
			totalPages = res.Pages
			domainList = make([]*Domain, 0, res.Total)
		}

		domainList = append(domainList, res.Embedded.Domains...)
		currentPage++
	}

	return domainList, nil
}

func (p *DNSProvider) getDomainInfo(domainID int) (*DomainResponse, error) {
	req, err := p.makeRequest(http.MethodGet, fmt.Sprintf("/v1/domains/%d", domainID), http.NoBody)
	if err != nil {
		return nil, err
	}

	var res DomainResponse
	err = p.sendRequest(req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (p *DNSProvider) getDomainIDByName(name string) (int, error) {
	// Load from cache if exists
	if id, ok := p.domainIDMapping[name]; ok {
		return id, nil
	}

	// Find out by querying API
	domains, err := p.listDomains()
	if err != nil {
		return -1, err
	}

	// Linear search over all registered domains
	for _, domain := range domains {
		if domain.Name == name || strings.HasSuffix(name, "."+domain.Name) {
			p.domainIDMapping[name] = domain.ID
			return domain.ID, nil
		}
	}

	return -1, nil
}

func (p *DNSProvider) getNameserverInfo(domainID int) (*NameserverResponse, error) {
	req, err := p.makeRequest(http.MethodGet, fmt.Sprintf("/v1/domains/%d/nameservers", domainID), http.NoBody)
	if err != nil {
		return nil, err
	}

	var res NameserverResponse
	if err := p.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (p *DNSProvider) isUsingCheckdomainNameservers(domainID int) (bool, error) {
	const (
		ns1 = "ns.checkdomain.de"
		ns2 = "ns2.checkdomain.de"
	)

	info, err := p.getNameserverInfo(domainID)
	if err != nil {
		return false, err
	}

	var (
		found1 = false
		found2 = false
	)
	for _, item := range info.Nameservers {
		if item.Name == ns1 {
			found1 = true
		} else if item.Name == ns2 {
			found2 = true
		}
	}

	return found1 && found2, nil
}

func (p *DNSProvider) listRecords(domainID int, recordType string) ([]*Record, error) {
	req, err := p.makeRequest(http.MethodGet, fmt.Sprintf("/v1/domains/%d/nameservers/records", domainID), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}

	q := req.URL.Query()
	q.Set("limit", strconv.Itoa(maxLimit))
	if recordType != "" {
		q.Set("type", recordType)
	}

	var (
		currentPage = 1
		totalPages  = maxInt
	)

	var recordList []*Record
	for currentPage <= totalPages {
		q.Set("page", strconv.Itoa(currentPage))
		req.URL.RawQuery = q.Encode()

		var res RecordListingResponse
		if err := p.sendRequest(req, &res); err != nil {
			return nil, fmt.Errorf("failed to send record listing request: %v", err)
		}

		// This is the first response, so we update
		// totalPages and allocate the slice memory.
		if totalPages == maxInt {
			totalPages = res.Pages
			recordList = make([]*Record, 0, res.Total)
		}

		recordList = append(recordList, res.Embedded.Records...)
		currentPage++
	}

	return recordList, nil
}

func (p *DNSProvider) replaceRecords(domainID int, records []*Record) error {
	bs, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("encoding record failed: %v", err)
	}

	req, err := p.makeRequest(http.MethodPut, fmt.Sprintf("/v1/domains/%d/nameservers/records", domainID), bytes.NewReader(bs))
	if err != nil {
		return err
	}

	return p.sendRequest(req, nil)
}

func (p *DNSProvider) createRecord(domainID int, record *Record) error {
	bs, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encoding record failed: %v", err)
	}

	req, err := p.makeRequest(http.MethodPost, fmt.Sprintf("/v1/domains/%d/nameservers/records", domainID), bytes.NewReader(bs))
	if err != nil {
		return err
	}

	return p.sendRequest(req, nil)
}

//nolint
func (p *DNSProvider) deleteRecord(domainID int, recordType, recordName, recordValue string) error {
	// Checkdomain doesn't seem provide a way to delete records
	// but one can replace all records at once.
	// The current solution is to fetch all records and
	// then use that list minus the record deleted as the new
	// record list.
	// TODO: Simplify this function once they do provide the functionality.

	domainInfo, err := p.getDomainInfo(domainID)
	if err != nil {
		return err
	}

	if strings.HasSuffix(recordName, "."+domainInfo.Name+".") {
		recordName = recordName[:len(recordName)-len(domainInfo.Name)-2]
	}

	nsInfo, err := p.getNameserverInfo(domainID)
	if err != nil {
		return err
	}

	allRecords, err := p.listRecords(domainID, "")
	if err != nil {
		return err
	}

	recordsToKeep := make([]*Record, 0, len(allRecords))
	// Find and delete matching records
	for _, record := range allRecords {
		var (
			typeMatch  = true
			nameMatch  = true
			valueMatch = true
		)

		if recordType != "" {
			typeMatch = record.Type == recordType
		}

		if recordName != "" {
			nameMatch = record.Name == recordName
		}

		if recordValue != "" {
			valueMatch = record.Value == recordValue
		}

		// Skip empty records
		if record.Value == "" {
			continue
		}

		// Skip our matching record
		if typeMatch && nameMatch && valueMatch {
			continue
		}

		// Skip some special records, otherwise we would get a "Nameserver update failed"
		if record.Type == "SOA" || record.Type == "NS" || record.Name == "@" || (nsInfo.General.IncludeWWW && record.Name == "www") {
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

	return p.replaceRecords(domainID, recordsToKeep)
}

func (p *DNSProvider) makeRequest(method, resource string, body io.Reader) (*http.Request, error) {
	uri, err := p.config.Endpoint.Parse(resource)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, uri.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.Token)
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (p *DNSProvider) sendRequest(req *http.Request, result interface{}) error {
	resp, err := p.httpClient.Do(req)
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

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return fmt.Errorf("unmarshaling %T error [status code=%d]: %v: %s", result, resp.StatusCode, err, string(raw))
	}
	return err
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	if resp.Body == nil {
		return fmt.Errorf("response body is nil, status code=%d", resp.StatusCode)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read body: status code=%d, error=%v", resp.StatusCode, err)
	}

	return fmt.Errorf("status code=%d: %s", resp.StatusCode, string(raw))
}
