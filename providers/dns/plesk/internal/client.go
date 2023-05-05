package internal

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// Client the Plesk API client.
type Client struct {
	login    string
	password string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient created a new Client.
func NewClient(baseURL *url.URL, login string, password string) *Client {
	return &Client{
		login:      login,
		password:   password,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetSite gets a site.
// https://docs.plesk.com/en-US/obsidian/api-rpc/about-xml-api/reference/managing-sites-domains/getting-information-about-sites.66583/
func (c Client) GetSite(ctx context.Context, domain string) (int, error) {
	payload := RequestPacketType{Site: &SiteTypeRequest{Get: SiteGetRequest{Filter: &SiteFilterType{
		Name: domain,
	}}}}

	response, err := c.doRequest(ctx, payload)
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
func (c Client) AddRecord(ctx context.Context, siteID int, host, value string) (int, error) {
	payload := RequestPacketType{DNS: &DNSInputType{AddRec: []AddRecRequest{{
		SiteID: siteID,
		Type:   "TXT",
		Host:   host,
		Value:  value,
	}}}}

	response, err := c.doRequest(ctx, payload)
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
func (c Client) DeleteRecord(ctx context.Context, recordID int) (int, error) {
	payload := RequestPacketType{DNS: &DNSInputType{DelRec: []DelRecRequest{{Filter: DNSSelectionFilterType{
		ID: recordID,
	}}}}}

	response, err := c.doRequest(ctx, payload)
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

func (c Client) doRequest(ctx context.Context, payload RequestPacketType) (*ResponsePacketType, error) {
	endpoint := c.baseURL.JoinPath("/enterprise/control/agent.php")

	body := new(bytes.Buffer)
	err := xml.NewEncoder(body).Encode(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml")

	req.Header.Set("Http_auth_login", c.login)
	req.Header.Set("Http_auth_passwd", c.password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return nil, errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var response ResponsePacketType
	err = xml.Unmarshal(raw, &response)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return &response, nil
}
