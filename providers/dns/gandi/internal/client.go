package internal

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// defaultBaseURL Gandi XML-RPC endpoint used by Present and CleanUp.
const defaultBaseURL = "https://rpc.gandi.net/xmlrpc/"

// Client the Gandi API client.
type Client struct {
	apiKey string

	BaseURL    string
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) GetZoneID(ctx context.Context, domain string) (int, error) {
	call := &methodCall{
		MethodName: "domain.info",
		Params: []param{
			paramString{Value: c.apiKey},
			paramString{Value: domain},
		},
	}

	resp := &responseStruct{}

	err := c.rpcCall(ctx, call, resp)
	if err != nil {
		return 0, err
	}

	var zoneID int
	for _, member := range resp.StructMembers {
		if member.Name == "zone_id" {
			zoneID = member.ValueInt
		}
	}

	if zoneID == 0 {
		return 0, fmt.Errorf("could not find zone_id for %s", domain)
	}
	return zoneID, nil
}

func (c *Client) CloneZone(ctx context.Context, zoneID int, name string) (int, error) {
	call := &methodCall{
		MethodName: "domain.zone.clone",
		Params: []param{
			paramString{Value: c.apiKey},
			paramInt{Value: zoneID},
			paramInt{Value: 0},
			paramStruct{
				StructMembers: []structMember{
					structMemberString{
						Name:  "name",
						Value: name,
					},
				},
			},
		},
	}

	resp := &responseStruct{}

	err := c.rpcCall(ctx, call, resp)
	if err != nil {
		return 0, err
	}

	var newZoneID int
	for _, member := range resp.StructMembers {
		if member.Name == "id" {
			newZoneID = member.ValueInt
		}
	}

	if newZoneID == 0 {
		return 0, errors.New("could not determine cloned zone_id")
	}
	return newZoneID, nil
}

func (c *Client) NewZoneVersion(ctx context.Context, zoneID int) (int, error) {
	call := &methodCall{
		MethodName: "domain.zone.version.new",
		Params: []param{
			paramString{Value: c.apiKey},
			paramInt{Value: zoneID},
		},
	}

	resp := &responseInt{}

	err := c.rpcCall(ctx, call, resp)
	if err != nil {
		return 0, err
	}

	if resp.Value == 0 {
		return 0, errors.New("could not create new zone version")
	}
	return resp.Value, nil
}

func (c *Client) AddTXTRecord(ctx context.Context, zoneID, version int, name, value string, ttl int) error {
	call := &methodCall{
		MethodName: "domain.zone.record.add",
		Params: []param{
			paramString{Value: c.apiKey},
			paramInt{Value: zoneID},
			paramInt{Value: version},
			paramStruct{
				StructMembers: []structMember{
					structMemberString{
						Name:  "type",
						Value: "TXT",
					}, structMemberString{
						Name:  "name",
						Value: name,
					}, structMemberString{
						Name:  "value",
						Value: value,
					}, structMemberInt{
						Name:  "ttl",
						Value: ttl,
					},
				},
			},
		},
	}

	resp := &responseStruct{}

	return c.rpcCall(ctx, call, resp)
}

func (c *Client) SetZoneVersion(ctx context.Context, zoneID, version int) error {
	call := &methodCall{
		MethodName: "domain.zone.version.set",
		Params: []param{
			paramString{Value: c.apiKey},
			paramInt{Value: zoneID},
			paramInt{Value: version},
		},
	}

	resp := &responseBool{}

	err := c.rpcCall(ctx, call, resp)
	if err != nil {
		return err
	}

	if !resp.Value {
		return errors.New("could not set zone version")
	}
	return nil
}

func (c *Client) SetZone(ctx context.Context, domain string, zoneID int) error {
	call := &methodCall{
		MethodName: "domain.zone.set",
		Params: []param{
			paramString{Value: c.apiKey},
			paramString{Value: domain},
			paramInt{Value: zoneID},
		},
	}

	resp := &responseStruct{}

	err := c.rpcCall(ctx, call, resp)
	if err != nil {
		return err
	}

	var respZoneID int
	for _, member := range resp.StructMembers {
		if member.Name == "zone_id" {
			respZoneID = member.ValueInt
		}
	}

	if respZoneID != zoneID {
		return fmt.Errorf("could not set new zone_id for %s", domain)
	}
	return nil
}

func (c *Client) DeleteZone(ctx context.Context, zoneID int) error {
	call := &methodCall{
		MethodName: "domain.zone.delete",
		Params: []param{
			paramString{Value: c.apiKey},
			paramInt{Value: zoneID},
		},
	}

	resp := &responseBool{}

	err := c.rpcCall(ctx, call, resp)
	if err != nil {
		return err
	}

	if !resp.Value {
		return errors.New("could not delete zone_id")
	}

	return nil
}

// rpcCall makes an XML-RPC call to Gandi's RPC endpoint by marshaling the data given in the call argument to XML
// and sending  that via HTTP Post to Gandi.
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
			FaultString: result.faultString(),
		}
	}

	return nil
}

func newXMLRequest(ctx context.Context, endpoint string, payload *methodCall) (*http.Request, error) {
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
