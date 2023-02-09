package pdns

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/miekg/dns"
)

type Record struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`

	// pre-v1 API
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  int    `json:"ttl,omitempty"`
}

type hostedZone struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	URL    string  `json:"url"`
	Kind   string  `json:"kind"`
	RRSets []rrSet `json:"rrsets"`

	// pre-v1 API
	Records []Record `json:"records"`
}

type rrSet struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Kind       string   `json:"kind"`
	ChangeType string   `json:"changetype"`
	Records    []Record `json:"records,omitempty"`
	TTL        int      `json:"ttl,omitempty"`
}

type rrSets struct {
	RRSets []rrSet `json:"rrsets"`
}

type apiError struct {
	ShortMsg string `json:"error"`
}

func (a apiError) Error() string {
	return a.ShortMsg
}

type apiVersion struct {
	URL     string `json:"url"`
	Version int    `json:"version"`
}

func (d *DNSProvider) getHostedZone(fqdn string) (*hostedZone, error) {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return nil, err
	}

	p := path.Join("/servers", d.config.ServerName, "/zones/", dns.Fqdn(authZone))

	result, err := d.sendRequest(http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	var zone hostedZone
	err = json.Unmarshal(result, &zone)
	if err != nil {
		return nil, err
	}

	// convert pre-v1 API result
	if len(zone.Records) > 0 {
		zone.RRSets = []rrSet{}
		for _, record := range zone.Records {
			set := rrSet{
				Name:    record.Name,
				Type:    record.Type,
				Records: []Record{record},
			}
			zone.RRSets = append(zone.RRSets, set)
		}
	}

	return &zone, nil
}

func (d *DNSProvider) findTxtRecord(fqdn string) (*rrSet, error) {
	zone, err := d.getHostedZone(fqdn)
	if err != nil {
		return nil, err
	}

	_, err = d.sendRequest(http.MethodGet, zone.URL, nil)
	if err != nil {
		return nil, err
	}

	for _, set := range zone.RRSets {
		if set.Type == "TXT" && (set.Name == dns01.UnFqdn(fqdn) || set.Name == fqdn) {
			return &set, nil
		}
	}

	return nil, nil
}

func (d *DNSProvider) getAPIVersion() (int, error) {
	result, err := d.sendRequest(http.MethodGet, "/api", nil)
	if err != nil {
		return 0, err
	}

	var versions []apiVersion
	err = json.Unmarshal(result, &versions)
	if err != nil {
		return 0, err
	}

	latestVersion := 0
	for _, v := range versions {
		if v.Version > latestVersion {
			latestVersion = v.Version
		}
	}

	return latestVersion, err
}

func (d *DNSProvider) notify(zone *hostedZone) error {
	if d.apiVersion < 1 || zone.Kind != "Master" && zone.Kind != "Slave" {
		return nil
	}

	_, err := d.sendRequest(http.MethodPut, path.Join(zone.URL, "/notify"), nil)
	if err != nil {
		return err
	}

	return nil
}

func (d *DNSProvider) sendRequest(method, uri string, body io.Reader) (json.RawMessage, error) {
	req, err := d.makeRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error talking to PDNS API: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		return nil, fmt.Errorf("unexpected HTTP status code %d when %sing '%s'", resp.StatusCode, req.Method, req.URL)
	}

	var msg json.RawMessage
	err = json.NewDecoder(resp.Body).Decode(&msg)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// empty body
			return nil, nil
		}
		// other error
		return nil, err
	}

	// check for PowerDNS error message
	if len(msg) > 0 && msg[0] == '{' {
		var errInfo apiError
		err = json.Unmarshal(msg, &errInfo)
		if err != nil {
			return nil, err
		}
		if errInfo.ShortMsg != "" {
			return nil, fmt.Errorf("error talking to PDNS API: %w", errInfo)
		}
	}
	return msg, nil
}

func (d *DNSProvider) makeRequest(method, uri string, body io.Reader) (*http.Request, error) {
	p := path.Join("/", uri)

	if p != "/api" && d.apiVersion > 0 && !strings.HasPrefix(p, "/api/v") {
		p = path.Join("/api", "v"+strconv.Itoa(d.apiVersion), p)
	}

	endpoint := d.config.Host.JoinPath(p)

	req, err := http.NewRequest(method, strings.TrimSuffix(endpoint.String(), "/"), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", d.config.APIKey)

	if method != http.MethodGet && method != http.MethodDelete {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
