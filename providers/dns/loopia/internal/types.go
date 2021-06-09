package internal

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// types for XML-RPC method calls and parameters

type param interface {
	param()
}

type paramString struct {
	XMLName xml.Name `xml:"param"`
	Value   string   `xml:"value>string"`
}

func (p paramString) param() {}

type paramInt struct {
	XMLName xml.Name `xml:"param"`
	Value   int      `xml:"value>int"`
}

func (p paramInt) param() {}

type paramStruct struct {
	XMLName       xml.Name       `xml:"param"`
	StructMembers []structMember `xml:"value>struct>member"`
}

func (p paramStruct) param() {}

type structMember interface {
	structMember()
}

type structMemberString struct {
	Name  string `xml:"name"`
	Value string `xml:"value>string"`
}

func (m structMemberString) structMember() {}

type structMemberInt struct {
	Name  string `xml:"name"`
	Value int    `xml:"value>int"`
}

func (m structMemberInt) structMember() {}

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

type responseString struct {
	responseFault
	Value string `xml:"params>param>value>string"`
}

type responseFault struct {
	FaultCode   int    `xml:"fault>value>struct>member>value>int"`
	FaultString string `xml:"fault>value>struct>member>value>string"`
}

func (r responseFault) faultCode() int      { return r.FaultCode }
func (r responseFault) faultString() string { return r.FaultString }

type rpcError struct {
	faultCode   int
	faultString string
}

func (e rpcError) Error() string {
	return fmt.Sprintf("RPC Error: (%d) %s", e.faultCode, e.faultString)
}

type recordObjectsResponse struct {
	responseFault
	XMLName xml.Name    `xml:"methodResponse"`
	Params  []RecordObj `xml:"params>param>value>array>data>value>struct"`
}

type RecordObj struct {
	Type     string
	TTL      int
	Priority int
	Rdata    string
	RecordID int
}

func (r *RecordObj) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
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

func (r *RecordObj) decodeValueString(name string, d *xml.Decoder, start xml.StartElement) error {
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

func (r *RecordObj) decodeValueInt(name string, d *xml.Decoder, start xml.StartElement) error {
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
