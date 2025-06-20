package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// defaultBaseURL for reaching the jSON-based API-Endpoint of netcup.
const defaultBaseURL = "https://ccp.netcup.net/run/webservice/servers/endpoint.php?JSON"

// Client netcup DNS client.
type Client struct {
	customerNumber string
	apiKey         string
	apiPassword    string

	baseURL    string
	HTTPClient *http.Client
}

// NewClient creates a netcup DNS client.
func NewClient(customerNumber, apiKey, apiPassword string) (*Client, error) {
	if customerNumber == "" || apiKey == "" || apiPassword == "" {
		return nil, errors.New("credentials missing")
	}

	return &Client{
		customerNumber: customerNumber,
		apiKey:         apiKey,
		apiPassword:    apiPassword,
		baseURL:        defaultBaseURL,
		HTTPClient:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// UpdateDNSRecord performs an update of the DNSRecords as specified by the netcup WSDL.
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (c *Client) UpdateDNSRecord(ctx context.Context, domainName string, records []DNSRecord) error {
	payload := &Request{
		Action: "updateDnsRecords",
		Param: UpdateDNSRecordsRequest{
			DomainName:      domainName,
			CustomerNumber:  c.customerNumber,
			APIKey:          c.apiKey,
			APISessionID:    getSessionID(ctx),
			ClientRequestID: "",
			DNSRecordSet:    DNSRecordSet{DNSRecords: records},
		},
	}

	err := c.doRequest(ctx, payload, nil)
	if err != nil {
		return fmt.Errorf("error when sending the request: %w", err)
	}

	return nil
}

// GetDNSRecords retrieves all dns records of an DNS-Zone as specified by the netcup WSDL
// returns an array of DNSRecords.
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (c *Client) GetDNSRecords(ctx context.Context, hostname string) ([]DNSRecord, error) {
	payload := &Request{
		Action: "infoDnsRecords",
		Param: InfoDNSRecordsRequest{
			DomainName:      hostname,
			CustomerNumber:  c.customerNumber,
			APIKey:          c.apiKey,
			APISessionID:    getSessionID(ctx),
			ClientRequestID: "",
		},
	}

	var responseData InfoDNSRecordsResponse
	err := c.doRequest(ctx, payload, &responseData)
	if err != nil {
		return nil, fmt.Errorf("error when sending the request: %w", err)
	}

	return responseData.DNSRecords, nil
}

// doRequest marshals given body to JSON, send the request to netcup API
// and returns body of response.
func (c *Client) doRequest(ctx context.Context, payload, result any) error {
	req, err := newJSONRequest(ctx, http.MethodPost, c.baseURL, payload)
	if err != nil {
		return err
	}

	req.Close = true

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusMultipleChoices {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	respMsg, err := unmarshalResponseMsg(req, resp)
	if err != nil {
		return err
	}

	if respMsg.Status != success {
		return respMsg
	}

	if result == nil {
		return nil
	}

	err = json.Unmarshal(respMsg.ResponseData, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, respMsg.ResponseData, err)
	}

	return nil
}

// GetDNSRecordIdx searches a given array of DNSRecords for a given DNSRecord
// equivalence is determined by Destination and RecortType attributes
// returns index of given DNSRecord in given array of DNSRecords.
func GetDNSRecordIdx(records []DNSRecord, record DNSRecord) (int, error) {
	for index, element := range records {
		if record.Destination == element.Destination && record.RecordType == element.RecordType {
			return index, nil
		}
	}
	return -1, errors.New("no DNS Record found")
}

func newJSONRequest(ctx context.Context, method, endpoint string, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func unmarshalResponseMsg(req *http.Request, resp *http.Response) (*ResponseMsg, error) {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var respMsg ResponseMsg
	err = json.Unmarshal(raw, &respMsg)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return &respMsg, nil
}
