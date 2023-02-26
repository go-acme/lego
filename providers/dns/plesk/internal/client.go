package internal

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client the Plesk API client.
type Client struct {
	HTTPClient *http.Client
	baseURL    *url.URL
	login      string
	password   string
}

// NewClient created a new Client.
func NewClient(baseURL *url.URL, login string, password string) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		login:      login,
		password:   password,
	}
}

// GetSite gets a site.
// https://docs.plesk.com/en-US/obsidian/api-rpc/about-xml-api/reference/managing-sites-domains/getting-information-about-sites.66583/
func (c Client) GetSite(domain string) (int, error) {
	payload := RequestPacketType{Site: &SiteTypeRequest{Get: SiteGetRequest{Filter: &SiteFilterType{
		Name: domain,
	}}}}

	response, err := c.do(payload)
	if err != nil {
		return 0, err
	}

	if response.System != nil {
		return 0, response.System
	}

	if response == nil || response.Site.Get.Result == nil {
		return 0, errors.New("unexpected empty result")
	}

	if response.Site.Get.Result.Status != StatusOK {
		return 0, response.Site.Get.Result
	}

	return response.Site.Get.Result.ID, nil
}

// AddRecord adds a TXT record.
// https://docs.plesk.com/en-US/obsidian/api-rpc/about-xml-api/reference/managing-dns/managing-dns-records/adding-dns-record.34798/
func (c Client) AddRecord(siteID int, host, value string) (int, error) {
	payload := RequestPacketType{DNS: &DNSInputType{AddRec: []AddRecRequest{{
		SiteID: siteID,
		Type:   "TXT",
		Host:   host,
		Value:  value,
	}}}}

	response, err := c.do(payload)
	if err != nil {
		return 0, err
	}

	if response.System != nil {
		return 0, response.System
	}

	if len(response.DNS.AddRec) < 1 {
		return 0, errors.New("unexpected empty result")
	}

	if response.DNS.AddRec[0].Result.Status != StatusOK {
		return 0, response.DNS.AddRec[0].Result
	}

	return response.DNS.AddRec[0].Result.ID, nil
}

// DeleteRecord Deletes a TXT record.
// https://docs.plesk.com/en-US/obsidian/api-rpc/about-xml-api/reference/managing-dns/managing-dns-records/deleting-dns-records.34864/
func (c Client) DeleteRecord(recordID int) (int, error) {
	payload := RequestPacketType{DNS: &DNSInputType{DelRec: []DelRecRequest{{Filter: DNSSelectionFilterType{
		ID: recordID,
	}}}}}

	response, err := c.do(payload)
	if err != nil {
		return 0, err
	}

	if response.System != nil {
		return 0, response.System
	}

	if len(response.DNS.DelRec) < 1 {
		return 0, errors.New("unexpected empty result")
	}

	if response.DNS.DelRec[0].Result.Status != StatusOK {
		return 0, response.DNS.DelRec[0].Result
	}

	return response.DNS.DelRec[0].Result.ID, nil
}

func (c Client) do(payload RequestPacketType) (*ResponsePacketType, error) {
	endpoint := c.baseURL.JoinPath("/enterprise/control/agent.php")

	body := &bytes.Buffer{}
	err := xml.NewEncoder(body).Encode(payload)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest(http.MethodPost, endpoint.String(), body)
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("Http_auth_login", c.login)
	req.Header.Set("Http_auth_passwd", c.password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		all, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(all))
	}

	var response ResponsePacketType
	err = xml.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
