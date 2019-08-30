package autodns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
)

type ResponseStatusCode string

const (
	ResponseStatusSuccess      ResponseStatusCode = `SUCCESS`
	ResponseStatusError                           = `ERROR`
	ResponseStatusNotify                          = `NOTIFY`
	ResponseStatusNotice                          = `NOTICE`
	ResponseStatusNiccomNotify                    = `NICCOM_NOTIFY`
)

type ResponseMessage struct {
	Text     string             `json:"text"`
	Messages []string           `json:"messages"`
	Objects  []string           `json:"objects"`
	Code     string             `json:"code"`
	Status   ResponseStatusCode `json:"status"`
}

type ResponseStatus struct {
	Code string             `json:"code"`
	Text string             `json:"text"`
	Type ResponseStatusCode `json:"type"`
}

type ResponseObject struct {
	Type    string `json:"type"`
	Value   string `json:"value"`
	Summary int32  `json:"summary"`
	Data    string
}

type DataZoneResponse struct {
	STID     string             `json:"stid"`
	CTID     string             `json:"ctid"`
	Messages []*ResponseMessage `json:"messages"`
	Status   *ResponseStatus    `json:"status"`
	Object   interface{}        `json:"object"`
	Data     []*Zone            `json:"data"`
}

// RecourceRecordTypeTXT is a txt record type
const RecourceRecordTypeTXT string = `TXT`

// ActionComplete defines the complete action
const ActionComplete string = `COMPLETE`

// ResourceRecord holds a resource record
type ResourceRecord struct {
	Name  string `json:"name"`
	TTL   int64  `json:"ttl"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Pref  int32  `json:"pref,omitempty"`
}

// Zone is an autodns zone record with all for us relevant fields
type Zone struct {
	Name              string            `json:"origin"`
	ResourceRecords   []*ResourceRecord `json:"resourceRecords"`
	Action            string            `json:"action"`
	VirtualNameServer string            `json:"virtualNameServer"`
}

// txtRecord holds a simplified version of a zone entry
type txtRecord struct {
	Name  string
	TTL   int64
	Value string
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

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Domainrobot-Context", strconv.Itoa(d.config.Context))
	req.SetBasicAuth(d.config.Username, d.config.Password)

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

	defer resp.Body.Close()

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

	defer resp.Body.Close()

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read body: status code=%d, error=%v", resp.StatusCode, err)
	}

	return fmt.Errorf("status code=%d: %s", resp.StatusCode, string(raw))
}

func (d *DNSProvider) makeZoneUpdateRequest(zone *Zone, domain string) (*Zone, error) {
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(zone); err != nil {
		return nil, err
	}

	req, err := d.makeRequest(http.MethodPut, path.Join("zone", domain, d.zoneNameservers[domain]), reqBody)
	if err != nil {
		return nil, err
	}

	var resp *Zone
	if err := d.sendRequest(req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *DNSProvider) addTxtRecord(domain, name, value string) (*Zone, error) {
	fmt.Println(domain, name, value)
	zone := &Zone{
		Name: name,
		ResourceRecords: []*ResourceRecord{
			{
				Name:  name,
				TTL:   120,
				Type:  "TXT",
				Value: value,
			},
		},
		Action: ActionComplete,
	}

	return d.makeZoneUpdateRequest(zone, domain)
}

func (d *DNSProvider) removeTxtRecord(domain, id string) error {
	zone := &Zone{
		Name:            domain,
		ResourceRecords: []*ResourceRecord{},
		Action:          ActionComplete,
	}

	_, err := d.makeZoneUpdateRequest(zone, domain)
	return err
}

func (d *DNSProvider) getTxtRecords(domain string) (*DataZoneResponse, error) {
	req, err := d.makeRequest(http.MethodGet, path.Join("zone", domain, d.zoneNameservers[domain]), nil)
	if err != nil {
		return nil, err
	}

	var resp *DataZoneResponse
	if err := d.sendRequest(req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
