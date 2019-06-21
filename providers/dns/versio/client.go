package versio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
)

const defaultBaseURL = "https://www.versio.nl/api/v1/"

type dnsRecordsResponse struct {
	Record dnsRecord `json:"domainInfo"`
}

type dnsRecord struct {
	DNSRecords []record `json:"dns_records"`
}

type record struct {
	Type  string `json:"type,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
	Prio  int    `json:"prio,omitempty"`
	TTL   int    `json:"ttl,omitempty"`
}

type dnsErrorResponse struct {
	Error errorMessage `json:"error"`
}

type errorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (d *DNSProvider) postDNSRecords(domain string, msg interface{}) error {

	reqBody := &bytes.Buffer{}
	err := json.NewEncoder(reqBody).Encode(msg)
	if err != nil {
		return err
	}

	newURI := path.Join(d.config.BaseURL.EscapedPath(), "domains/"+domain+"/update")
	endpoint, err := d.config.BaseURL.Parse(newURI)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint.String(), reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if len(d.config.Username) > 0 && len(d.config.Password) > 0 {
		req.SetBasicAuth(d.config.Username, d.config.Password)
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, berr := ioutil.ReadAll(resp.Body)
		if berr != nil {
			return fmt.Errorf("%d: failed to read response body: %v", resp.StatusCode, err)
		}

		respError := &dnsErrorResponse{}
		_ = json.Unmarshal(body, respError)
		return fmt.Errorf("%d: request failed: %v", resp.StatusCode, respError.Error.Message)
	}

	return nil
}

func (d *DNSProvider) getDNSRecords(domain string) (*dnsRecordsResponse, error) {
	reqBody := &bytes.Buffer{}

	newURI := path.Join(d.config.BaseURL.EscapedPath(), "domains/"+domain+"?show_dns_records=true")
	endpoint, err := d.config.BaseURL.Parse(newURI)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if len(d.config.Username) > 0 && len(d.config.Password) > 0 {
		req.SetBasicAuth(d.config.Username, d.config.Password)
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, berr := ioutil.ReadAll(resp.Body)
		if berr != nil {
			return nil, fmt.Errorf("%d: failed to read response body: %v", resp.StatusCode, err)
		}

		respError := &dnsErrorResponse{}
		_ = json.Unmarshal(body, respError)
		return nil, fmt.Errorf("%d: request failed: %v", resp.StatusCode, respError.Error.Message)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("request failed: Body: %s Err: %v", string(content), err)
	}
	// Everything looks good; we'll need all the dns_records to add the new TXT record
	respData := &dnsRecordsResponse{}
	err = json.Unmarshal(content, respData)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, content)
	}

	return respData, nil

}
