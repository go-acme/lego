package safedns

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
)

const defaultBaseURL = "https://api.ukfast.io"

type txtResponse struct {
	Data struct {
		ID int `json:"id"`
	} `json:"data"`
	Meta struct {
		Location string `json:"location"`
	}
}

type recordType string

const (
	typeTXT recordType = "TXT"
)

type record struct {
	Name    string     `json:"name"`
	Type    recordType `json:"type"`
	Content string     `json:"content"`
	TTL     int        `json:"ttl"`
}

type apiError struct {
	Message string `json:"message"`
}

func (d *DNSProvider) addTxtRecord(fqdn, value string) (*txtResponse, error) {
	zone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(fqdn))
	if err != nil {
		return nil, fmt.Errorf("could not determine zone for domain: %q: %w", fqdn, err)
	}

	reqData := record{Name: dns01.UnFqdn(fqdn), Type: typeTXT, Content: fmt.Sprintf("%q", value), TTL: d.config.TTL}
	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/safedns/v1/zones/%s/records", d.config.BaseURL, dns01.UnFqdn(zone))
	req, err := d.newRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	log.Infof("safedns: creating record %+v at %s", reqData, url)

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, readError(req, resp)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(toUnreadableBodyMessage(req, content))
	}

	respData := &txtResponse{}
	err = json.Unmarshal(content, respData)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, toUnreadableBodyMessage(req, content))
	}

	log.Infof("safedns: created record with ID %d", respData.Data.ID)

	return respData, nil
}

func (d *DNSProvider) removeTxtRecord(domain string, recordID int) error {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("safedns: could not determine zone for domain %q: %w", domain, err)
	}

	reqURL := fmt.Sprintf("%s/safedns/v1/zones/%s/records/%d", d.config.BaseURL, dns01.UnFqdn(authZone), recordID)
	req, err := d.newRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	log.Infof("safedns: cleaning up record %d at %s", recordID, reqURL)

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return readError(req, resp)
	}

	return nil
}

func (d *DNSProvider) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", d.config.AuthToken)

	return req, nil
}

func readError(req *http.Request, resp *http.Response) error {
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New(toUnreadableBodyMessage(req, content))
	}

	var errInfo apiError
	err = json.Unmarshal(content, &errInfo)
	if err != nil {
		return fmt.Errorf("safedns: unmarshaling error: %w: %s", err, toUnreadableBodyMessage(req, content))
	}

	return fmt.Errorf("safedns: HTTP %d: %s", resp.StatusCode, content)
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s received a response with an invalid format: %q", req.URL, string(rawBody))
}
