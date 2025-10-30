package internal

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const (
	apiBaseURL = "https://api.nic.ru/dns-master"
	tokenURL   = "https://api.nic.ru/oauth/token"
)

const successStatus = "success"

// Trimmer trim all XML fields.
type Trimmer struct {
	decoder *xml.Decoder
}

func (tr Trimmer) Token() (xml.Token, error) {
	t, err := tr.decoder.Token()
	if cd, ok := t.(xml.CharData); ok {
		t = xml.CharData(bytes.TrimSpace(cd))
	}

	return t, err
}

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	baseURL, _ := url.Parse(apiBaseURL)

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}, nil
}

func (c *Client) GetServices(ctx context.Context) ([]Service, error) {
	endpoint := c.baseURL.JoinPath("services")

	req, err := newXMLRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	apiResponse, err := c.do(req)
	if err != nil {
		return nil, err
	}

	if apiResponse.Data == nil {
		return nil, nil
	}

	return apiResponse.Data.Service, nil
}

func (c *Client) ListZones(ctx context.Context) ([]Zone, error) {
	endpoint := c.baseURL.JoinPath("zones")

	req, err := newXMLRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	apiResponse, err := c.do(req)
	if err != nil {
		return nil, err
	}

	if apiResponse.Data == nil {
		return nil, nil
	}

	return apiResponse.Data.Zone, nil
}

func (c *Client) GetZonesByService(ctx context.Context, serviceName string) ([]Zone, error) {
	endpoint := c.baseURL.JoinPath("services", serviceName, "zones")

	req, err := newXMLRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	apiResponse, err := c.do(req)
	if err != nil {
		return nil, err
	}

	if apiResponse.Data == nil {
		return nil, nil
	}

	return apiResponse.Data.Zone, nil
}

func (c *Client) GetRecords(ctx context.Context, serviceName, zoneName string) ([]RR, error) {
	endpoint := c.baseURL.JoinPath("services", serviceName, "zones", zoneName, "records")

	req, err := newXMLRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	apiResponse, err := c.do(req)
	if err != nil {
		return nil, err
	}

	if apiResponse.Data == nil {
		return nil, nil
	}

	var records []RR
	for _, zone := range apiResponse.Data.Zone {
		records = append(records, zone.RR...)
	}

	return records, nil
}

func (c *Client) DeleteRecord(ctx context.Context, serviceName, zoneName, id string) error {
	endpoint := c.baseURL.JoinPath("services", serviceName, "zones", zoneName, "records", id)

	req, err := newXMLRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CommitZone(ctx context.Context, serviceName, zoneName string) error {
	endpoint := c.baseURL.JoinPath("services", serviceName, "zones", zoneName, "commit")

	req, err := newXMLRequest(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) AddRecords(ctx context.Context, serviceName, zoneName string, rrs []RR) ([]Zone, error) {
	endpoint := c.baseURL.JoinPath("services", serviceName, "zones", zoneName, "records")

	payload := &Request{RRList: &RRList{RR: rrs}}

	req, err := newXMLRequest(ctx, http.MethodPut, endpoint, payload)
	if err != nil {
		return nil, err
	}

	apiResponse, err := c.do(req)
	if err != nil {
		return nil, err
	}

	if apiResponse.Data == nil {
		return nil, nil
	}

	return apiResponse.Data.Zone, nil
}

func (c *Client) do(req *http.Request) (*Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	apiResponse := &Response{}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	decoder := xml.NewTokenDecoder(Trimmer{decoder: xml.NewDecoder(bytes.NewReader(raw))})

	err = decoder.Decode(apiResponse)
	if err != nil {
		return nil, fmt.Errorf("[status code=%d] decode XML response: %s", resp.StatusCode, string(raw))
	}

	if apiResponse.Status != successStatus {
		return nil, fmt.Errorf("[status code=%d] %s: %w", resp.StatusCode, apiResponse.Status, apiResponse.Errors.Error)
	}

	return apiResponse, nil
}

func newXMLRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	body := new(bytes.Buffer)

	if payload != nil {
		body.WriteString(xml.Header)

		encoder := xml.NewEncoder(body)
		encoder.Indent("", "  ")

		err := encoder.Encode(payload)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "text/xml")

	if payload != nil {
		req.Header.Set("Content-Type", "text/xml")
	}

	return req, nil
}
