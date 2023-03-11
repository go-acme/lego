package internal

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// DefaultBaseURL is url to the XML-RPC api.
const DefaultBaseURL = "https://api.loopia.se/RPCSERV"

// Client the Loopia client.
type Client struct {
	apiUser     string
	apiPassword string

	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Loopia Client.
func NewClient(apiUser, apiPassword string) *Client {
	return &Client{
		apiUser:     apiUser,
		apiPassword: apiPassword,
		BaseURL:     DefaultBaseURL,
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// AddTXTRecord adds a TXT record.
func (c *Client) AddTXTRecord(ctx context.Context, domain string, subdomain string, ttl int, value string) error {
	call := &methodCall{
		MethodName: "addZoneRecord",
		Params: []param{
			paramString{Value: c.apiUser},
			paramString{Value: c.apiPassword},
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

	err := c.rpcCall(ctx, call, resp)
	if err != nil {
		return err
	}

	return checkResponse(resp.Value)
}

// RemoveTXTRecord removes a TXT record.
func (c *Client) RemoveTXTRecord(ctx context.Context, domain string, subdomain string, recordID int) error {
	call := &methodCall{
		MethodName: "removeZoneRecord",
		Params: []param{
			paramString{Value: c.apiUser},
			paramString{Value: c.apiPassword},
			paramString{Value: domain},
			paramString{Value: subdomain},
			paramInt{Value: recordID},
		},
	}
	resp := &responseString{}

	err := c.rpcCall(ctx, call, resp)
	if err != nil {
		return err
	}

	return checkResponse(resp.Value)
}

// GetTXTRecords gets TXT records.
func (c *Client) GetTXTRecords(ctx context.Context, domain string, subdomain string) ([]RecordObj, error) {
	call := &methodCall{
		MethodName: "getZoneRecords",
		Params: []param{
			paramString{Value: c.apiUser},
			paramString{Value: c.apiPassword},
			paramString{Value: domain},
			paramString{Value: subdomain},
		},
	}
	resp := &recordObjectsResponse{}

	err := c.rpcCall(ctx, call, resp)

	return resp.Params, err
}

// RemoveSubdomain remove a sub-domain.
func (c *Client) RemoveSubdomain(ctx context.Context, domain, subdomain string) error {
	call := &methodCall{
		MethodName: "removeSubdomain",
		Params: []param{
			paramString{Value: c.apiUser},
			paramString{Value: c.apiPassword},
			paramString{Value: domain},
			paramString{Value: subdomain},
		},
	}
	resp := &responseString{}

	err := c.rpcCall(ctx, call, resp)
	if err != nil {
		return err
	}

	return checkResponse(resp.Value)
}

// rpcCall makes an XML-RPC call to Loopia's RPC endpoint by marshaling the data given in the call argument to XML
// and sending that via HTTP Post to Loopia.
// The response is then unmarshalled into the resp argument.
func (c *Client) rpcCall(ctx context.Context, call *methodCall, result response) error {
	req, err := newXMLRequest(ctx, c.BaseURL, call)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = xml.Unmarshal(raw, result)
	if err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	if result.faultCode() != 0 {
		return RPCError{
			FaultCode:   result.faultCode(),
			FaultString: strings.TrimSpace(result.faultString()),
		}
	}

	return nil
}

func newXMLRequest(ctx context.Context, endpoint string, payload any) (*http.Request, error) {
	body := new(bytes.Buffer)
	body.WriteString(xml.Header)

	encoder := xml.NewEncoder(body)
	encoder.Indent("", "  ")

	err := encoder.Encode(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml")

	return req, nil
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
