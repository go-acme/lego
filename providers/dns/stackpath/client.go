package stackpath

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/xenolf/lego/acme"
)

// Zone is the item response from the Stackpath api getZone
type Zone struct {
	ID     string
	Domain string
}

// GetZoneResponse is the response struct from the Stackpath api getZone
type GetZoneResponse struct {
	Zones []*Zone
}

// GetRecordsResponse is the response struct from the Stackpath api GetRecords
type GetRecordsResponse struct {
	ZoneRecords []*Record
}

// Record is the item response from the Stackpath api GetRecords
type Record struct {
	ID   string
	Name string
	Type string
}

func (d *DNSProvider) httpGet(path string, in interface{}) error {
	resp, err := d.client.Get(fmt.Sprintf("%s/%s%s", defaultBaseURL, d.config.StackID, path))
	if err != nil {
		return err
	}

	rawBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("non 200 response: %d - %s", resp.StatusCode, string(rawBody))
	}

	if err := json.Unmarshal(rawBody, in); err != nil {
		return err
	}

	return nil
}

func (d *DNSProvider) httpPost(path string, body interface{}) error {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := d.client.Post(
		fmt.Sprintf("%s/%s%s", defaultBaseURL, d.config.StackID, path),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return err
	}

	rawBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("non 200 response: %d - %s", resp.StatusCode, string(rawBody))
	}

	return nil
}

func (d *DNSProvider) httpDelete(path string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s%s", defaultBaseURL, d.config.StackID, path), nil)
	if err != nil {
		return err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}

	rawBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("non 200 response: %d - %s", resp.StatusCode, string(rawBody))
	}

	return nil
}

func (d *DNSProvider) getZoneForDomain(domain string) (*Zone, error) {
	domain = acme.UnFqdn(domain)

	var zones GetZoneResponse
	if err := d.httpGet(fmt.Sprintf("/zones?page_request.filter=domain = '%s'", domain), &zones); err != nil {
		return nil, err
	}

	if len(zones.Zones) == 0 {
		return nil, fmt.Errorf("did not find zone with domain %s", domain)
	}

	return zones.Zones[0], nil
}

func (d *DNSProvider) getRecordForZone(name string, zone *Zone) (*Record, error) {
	var records GetRecordsResponse
	err := d.httpGet(
		fmt.Sprintf("/zones%s/records?page_request.filter=name = '%s' and type = 'TXT'", zone.ID, name),
		&records,
	)
	if err != nil {
		return nil, err
	}

	if len(records.ZoneRecords) == 0 {
		return nil, fmt.Errorf("did not find record with name %s", name)
	}

	return records.ZoneRecords[0], nil
}
