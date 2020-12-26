package loopia

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	defaultBaseURL = "https://api.loopia.se/RPCSERV"
	minTTL         = 300
)

const (
	returnOk        = "OK"
	returnAuthError = "AUTH_ERROR"
)

type dnsClient interface {
	addTXTRecord(string, string, int, string) error
	removeTXTRecord(string, string, int) error
	getTXTRecords(string, string) ([]recordObj, error)
	removeSubdomain(string, string) error
}

// Client Loopia client.
type Client struct {
	APIUser     string
	APIPassword string
	BaseURL     string
	HTTPClient  *http.Client
}

// NewClient creates a new Loopia Client.
func NewClient(apiUser, apiPassword string) Client {
	return Client{
		APIUser:     apiUser,
		APIPassword: apiPassword,
		BaseURL:     defaultBaseURL,
		HTTPClient:  &http.Client{},
	}
}

// types for XML-RPC method calls and parameters

type param interface {
	param()
}

type paramString struct {
	XMLName xml.Name `xml:"param"`
	Value   string   `xml:"value>string"`
}

type paramInt struct {
	XMLName xml.Name `xml:"param"`
	Value   int      `xml:"value>int"`
}

type structMember interface {
	structMember()
}

type structMemberString struct {
	Name  string `xml:"name"`
	Value string `xml:"value>string"`
}

type structMemberInt struct {
	Name  string `xml:"name"`
	Value int    `xml:"value>int"`
}

type paramStruct struct {
	XMLName       xml.Name       `xml:"param"`
	StructMembers []structMember `xml:"value>struct>member"`
}

func (p paramString) param()               {}
func (p paramInt) param()                  {}
func (m structMemberString) structMember() {}
func (m structMemberInt) structMember()    {}
func (p paramStruct) param()               {}

type methodCall struct {
	XMLName    xml.Name `xml:"methodCall"`
	MethodName string   `xml:"methodName"`
	Params     []param  `xml:"params>param"`
}

// types for XML-RPC responses

type response interface {
	faultCode() int
	faultString() string
}

type responseFault struct {
	FaultCode   int    `xml:"fault>value>struct>member>value>int"`
	FaultString string `xml:"fault>value>struct>member>value>string"`
}

func (r responseFault) faultCode() int      { return r.FaultCode }
func (r responseFault) faultString() string { return r.FaultString }

type responseString struct {
	responseFault
	Value string `xml:"params>param>value>string"`
}

type rpcError struct {
	faultCode   int
	faultString string
}

type recordObjectsResponse struct {
	responseFault
	XMLName xml.Name    `xml:"methodResponse"`
	Params  []recordObj `xml:"params>param>value>array>data>value>struct"`
}

type recordObj struct {
	Type     string
	TTL      int
	Priority int
	Rdata    string
	RecordID int
}

func (r *recordObj) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var name string
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}
		switch tt := t.(type) {
		case xml.StartElement:
			switch tt.Name.Local {
			case "name": // The name of the record object: <name>
				var s string
				if err = d.DecodeElement(&s, &start); err != nil {
					return err
				}
				name = strings.TrimSpace(s)
			case "string": // A string value of the record object: <value><string>
				if err = r.decodeValueString(name, d, start); err != nil {
					return err
				}
			case "int": // An int value of the record object: <value><int>
				if err = r.decodeValueInt(name, d, start); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}
	}
}

func (r *recordObj) decodeValueString(name string, d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	s = strings.TrimSpace(s)
	switch name {
	case "type":
		r.Type = s
	case "rdata":
		r.Rdata = s
	}
	return nil
}

func (r *recordObj) decodeValueInt(name string, d *xml.Decoder, start xml.StartElement) error {
	var i int
	if err := d.DecodeElement(&i, &start); err != nil {
		return err
	}
	switch name {
	case "record_id":
		r.RecordID = i
	case "ttl":
		r.TTL = i
	case "priority":
		r.Priority = i
	}
	return nil
}

func (e rpcError) Error() string {
	return fmt.Sprintf("Loopia DNS: RPC Error: (%d) %s", e.faultCode, e.faultString)
}

// rpcCall makes an XML-RPC call to loopia's RPC endpoint by
// marshaling the data given in the call argument to XML and sending
// that via HTTP Post to loopia.
// The response is then unmarshalled into the resp argument.
func (c *Client) rpcCall(call *methodCall, resp response) error {
	// marshal
	b, err := xml.MarshalIndent(call, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	// post
	b = append([]byte(`<?xml version="1.0"?>`+"\n"), b...)
	respBody, err := c.httpPost(c.BaseURL, "text/xml", bytes.NewReader(b))
	if err != nil {
		return err
	}

	// unmarshal
	err = xml.Unmarshal(respBody, resp)
	if err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Post Error: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTTP Post Error: %w", err)
	}

	return b, nil
}

func (c *Client) addTXTRecord(domain string, subdomain string, ttl int, value string) error {
	resp := &responseString{}
	err := c.rpcCall(&methodCall{
		MethodName: "addZoneRecord",
		Params: []param{
			paramString{Value: c.APIUser},
			paramString{Value: c.APIPassword},
			paramString{Value: domain},
			paramString{Value: subdomain},
			paramStruct{
				StructMembers: []structMember{
					structMemberString{
						Name:  "type",
						Value: "TXT",
					}, structMemberInt{
						Name:  "ttl",
						Value: ttl,
					}, structMemberInt{
						Name:  "priority",
						Value: 0,
					}, structMemberString{
						Name:  "rdata",
						Value: value,
					}, structMemberInt{
						Name:  "record_id",
						Value: 0,
					},
				},
			},
		},
	}, resp)
	if err != nil {
		return err
	}
	switch v := strings.TrimSpace(resp.Value); v {
	case returnOk:
		return nil
	case returnAuthError:
		return fmt.Errorf("Authentication Error")
	default:
		return fmt.Errorf("Unknown Error: '%s'", v)
	}
}

func (c *Client) removeTXTRecord(domain string, subdomain string, recordID int) error {
	resp := &responseString{}
	err := c.rpcCall(&methodCall{
		MethodName: "removeZoneRecord",
		Params: []param{
			paramString{Value: c.APIUser},
			paramString{Value: c.APIPassword},
			paramString{Value: domain},
			paramString{Value: subdomain},
			paramInt{Value: recordID},
		},
	}, resp)
	if err != nil {
		return err
	}
	switch v := strings.TrimSpace(resp.Value); v {
	case returnOk:
		return nil
	case returnAuthError:
		return fmt.Errorf("Authentication Error")
	default:
		return fmt.Errorf("Unknown Error: '%s'", v)
	}
}

func (c *Client) getTXTRecords(domain string, subdomain string) ([]recordObj, error) {
	resp := &recordObjectsResponse{}

	err := c.rpcCall(&methodCall{
		MethodName: "getZoneRecords",
		Params: []param{
			paramString{Value: c.APIUser},
			paramString{Value: c.APIPassword},
			paramString{Value: domain},
			paramString{Value: subdomain},
		},
	}, resp)

	return resp.Params, err
}

func (c *Client) removeSubdomain(domain, subdomain string) error {
	resp := &responseString{}
	err := c.rpcCall(&methodCall{
		MethodName: "removeSubdomain",
		Params: []param{
			paramString{Value: c.APIUser},
			paramString{Value: c.APIPassword},
			paramString{Value: domain},
			paramString{Value: subdomain},
		},
	}, resp)
	if err != nil {
		return err
	}
	switch v := strings.TrimSpace(resp.Value); v {
	case returnOk:
		return nil
	case returnAuthError:
		return fmt.Errorf("Authentication Error")
	default:
		return fmt.Errorf("Unknown Error: '%s'", v)
	}
}
