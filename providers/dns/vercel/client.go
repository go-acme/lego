package vercel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"io"
	"net/http"
)

const defaultBaseURL = "https://api.vercel.com"

// txtRecordResponse represents a response from Vercel's API after making a TXT record.
type txtRecordResponse struct {
	UID     string `json:"uid"`
	Updated int    `json:"updated,omitempty"`
}

type record struct {
	ID    string `json:"id,omitempty"`
	Slug  string `json:"slug,omitempty"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
	TTL   int    `json:"ttl,omitempty"`
}

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (d *DNSProvider) removeTxtRecord(domain string, recordID string) error {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("could not determine zone for domain %q: %w", domain, err)
	}

	reqURL := fmt.Sprintf("%s/v2/domains/%s/records/%s", d.config.BaseURL, dns01.UnFqdn(authZone), recordID)
	req, err := d.newRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	resp, err := d.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return readError(req, resp)
	}

	return nil
}

func (d *DNSProvider) addTxtRecord(fqdn, value string) (*txtRecordResponse, error) {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(fqdn))
	if err != nil {
		return nil, fmt.Errorf("could not determine zone for domain %q: %w", fqdn, err)
	}

	reqData := record{Type: "TXT", Name: fqdn, Value: value, TTL: d.config.TTL}
	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/v2/domains/%s/records", d.config.BaseURL, dns01.UnFqdn(authZone))
	req, err := d.newRequest(http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := d.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, readError(req, resp)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(toUnreadableBodyMessage(req, content))
	}

	// Everything looks good; but we'll need the ID later to delete the record
	respData := &txtRecordResponse{}
	err = json.Unmarshal(content, respData)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, toUnreadableBodyMessage(req, content))
	}

	return respData, nil
}

func (d *DNSProvider) newRequest(method, reqURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.config.AuthToken))

	if d.config.TeamID != "" {
		query := req.URL.Query()
		query.Add("teamId", d.config.TeamID)
		req.URL.RawQuery = query.Encode()
	}

	return req, nil
}

func (d *DNSProvider) httpClient() *http.Client {
	if d.config.HTTPClient != nil {
		return d.config.HTTPClient
	}

	return http.DefaultClient
}

func readError(req *http.Request, resp *http.Response) error {
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New(toUnreadableBodyMessage(req, content))
	}

	var errInfo apiError
	err = json.Unmarshal(content, &errInfo)
	if err != nil {
		return fmt.Errorf("apiError unmarshaling error: %w: %s", err, toUnreadableBodyMessage(req, content))
	}

	return fmt.Errorf("HTTP %d: %s: %s", resp.StatusCode, errInfo.Error.Code, errInfo.Error.Message)
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", req.URL, string(rawBody))
}
