package stackpath

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/xenolf/lego/acme"
	"golang.org/x/net/publicsuffix"
)

// Zone is the item response from the Stackpath api getZone
type Zone struct {
	ID     string
	Domain string
}

// getZoneResponse is the response struct from the Stackpath api getZone
type getZoneResponse struct {
	Zones []*Zone
}

// getRecordsResponse is the response struct from the Stackpath api GetRecords
type getRecordsResponse struct {
	Records []*Record
}

// Record is the item response from the Stackpath api GetRecords
type Record struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  int    `json:"ttl"`
	Data string `json:"data"`
}

func (d *DNSProvider) httpGet(path string, in interface{}) error {
	resp, err := d.client.Get(fmt.Sprintf("%s/%s%s", defaultBaseURL, d.config.StackID, path))
	if err != nil {
		return err
	}

	if resp.Body == nil {
		return fmt.Errorf("no response body for request GET %s", path)
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

	if resp.StatusCode > 299 {
		if resp.Body == nil {
			return fmt.Errorf("no response body for request POST %s", path)
		}

		rawBody, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return err
		}

		return fmt.Errorf("non 200 response: %d - %s", resp.StatusCode, string(rawBody))
	}

	return nil
}

func (d *DNSProvider) httpDelete(path string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/%s%s", defaultBaseURL, d.config.StackID, path),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode > 299 {
		if resp.Body == nil {
			return fmt.Errorf("no response body for request DELETES %s", path)
		}

		rawBody, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return err
		}

		return fmt.Errorf("non 200 response: %d - %s", resp.StatusCode, string(rawBody))
	}

	return nil
}

func (d *DNSProvider) getZoneForDomain(domain string) (*Zone, error) {
	domain = acme.UnFqdn(domain)
	tld, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("page_request.filter", fmt.Sprintf("domain='%s'", tld))

	var zones getZoneResponse
	err = d.httpGet(fmt.Sprintf("/zones?%s", params.Encode()), &zones)
	if err != nil {
		return nil, err
	}

	if len(zones.Zones) == 0 {
		return nil, fmt.Errorf("did not find zone with domain %s", domain)
	}

	return zones.Zones[0], nil
}

func (d *DNSProvider) getRecordForZone(name string, zone *Zone) (*Record, error) {
	params := url.Values{}
	params.Add("page_request.filter", fmt.Sprintf("name='%s' and type='TXT'", name))

	var records getRecordsResponse
	err := d.httpGet(fmt.Sprintf("/zones/%s/records?%s", zone.ID, params.Encode()), &records)
	if err != nil {
		return nil, err
	}

	if len(records.Records) == 0 {
		return nil, fmt.Errorf("did not find record with name %s", name)
	}

	return records.Records[0], nil
}
