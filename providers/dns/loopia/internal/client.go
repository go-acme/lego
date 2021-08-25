package internal

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultBaseURL is url to the XML-RPC api.
const DefaultBaseURL = "https://api.loopia.se/RPCSERV"

// Client the Loopia client.
type Client struct {
	APIUser     string
	APIPassword string
	BaseURL     string
	HTTPClient  *http.Client
}

// NewClient creates a new Loopia Client.
func NewClient(apiUser, apiPassword string) *Client {
	return &Client{
		APIUser:     apiUser,
		APIPassword: apiPassword,
		BaseURL:     DefaultBaseURL,
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// AddTXTRecord adds a TXT record.
func (c *Client) AddTXTRecord(domain string, subdomain string, ttl int, value string) error {
	call := &methodCall{
		MethodName: "addZoneRecord",
		Params: []param{
			paramString{Value: c.APIUser},
			paramString{Value: c.APIPassword},
			paramString{Value: domain},
			paramString{Value: subdomain},
			paramStruct{
				StructMembers: []structMember{
					structMemberString{Name: "type", Value: "TXT"},
					structMemberInt{Name: "ttl", Value: ttl},
					structMemberInt{Name: "priority", Value: 0},
					structMemberString{Name: "rdata", Value: value},
					structMemberInt{Name: "record_id", Value: 0},
				},
			},
		},
	}
	resp := &responseString{}

	err := c.rpcCall(call, resp)
	if err != nil {
		return err
	}

	return checkResponse(resp.Value)
}

// RemoveTXTRecord removes a TXT record.
func (c *Client) RemoveTXTRecord(domain string, subdomain string, recordID int) error {
	call := &methodCall{
		MethodName: "removeZoneRecord",
		Params: []param{
			paramString{Value: c.APIUser},
			paramString{Value: c.APIPassword},
			paramString{Value: domain},
			paramString{Value: subdomain},
			paramInt{Value: recordID},
		},
	}
	resp := &responseString{}

	err := c.rpcCall(call, resp)
	if err != nil {
		return err
	}

	return checkResponse(resp.Value)
}

// GetTXTRecords gets TXT records.
func (c *Client) GetTXTRecords(domain string, subdomain string) ([]RecordObj, error) {
	call := &methodCall{
		MethodName: "getZoneRecords",
		Params: []param{
			paramString{Value: c.APIUser},
			paramString{Value: c.APIPassword},
			paramString{Value: domain},
			paramString{Value: subdomain},
		},
	}
	resp := &recordObjectsResponse{}

	err := c.rpcCall(call, resp)

	return resp.Params, err
}

// RemoveSubdomain remove a sub-domain.
func (c *Client) RemoveSubdomain(domain, subdomain string) error {
	call := &methodCall{
		MethodName: "removeSubdomain",
		Params: []param{
			paramString{Value: c.APIUser},
			paramString{Value: c.APIPassword},
			paramString{Value: domain},
			paramString{Value: subdomain},
		},
	}
	resp := &responseString{}

	err := c.rpcCall(call, resp)
	if err != nil {
		return err
	}

	return checkResponse(resp.Value)
}

// rpcCall makes an XML-RPC call to Loopia's RPC endpoint
// by marshaling the data given in the call argument to XML and sending that via HTTP Post to Loopia.
// The response is then unmarshalled into the resp argument.
func (c *Client) rpcCall(call *methodCall, resp response) error {
	body, err := xml.MarshalIndent(call, "", "  ")
	if err != nil {
		return fmt.Errorf("error during unmarshalling the request body: %w", err)
	}

	body = append([]byte(`<?xml version="1.0"?>`+"\n"), body...)

	respBody, err := c.httpPost(c.BaseURL, "text/xml", bytes.NewReader(body))
	if err != nil {
		return err
	}

	err = xml.Unmarshal(respBody, resp)
	if err != nil {
		return fmt.Errorf("error during unmarshalling the response body: %w", err)
	}

	if resp.faultCode() != 0 {
		return rpcError{
			faultCode:   resp.faultCode(),
			faultString: strings.TrimSpace(resp.faultString()),
		}
	}

	return nil
}

func (c *Client) httpPost(url string, bodyType string, body io.Reader) ([]byte, error) {
	resp, err := c.HTTPClient.Post(url, bodyType, body)
	if err != nil {
		return nil, fmt.Errorf("HTTP Post Error: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Post Error: %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTTP Post Error: %w", err)
	}

	return b, nil
}

func checkResponse(value string) error {
	switch v := strings.TrimSpace(value); v {
	case "OK":
		return nil
	case "AUTH_ERROR":
		return errors.New("authentication error")
	default:
		return fmt.Errorf("unknown error: %q", v)
	}
}
