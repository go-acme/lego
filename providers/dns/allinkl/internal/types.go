package internal

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

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

// Fault a SOAP fault.
type Fault struct {
	Code    string `xml:"faultcode"`
	Message string `xml:"faultstring"`
	Actor   string `xml:"faultactor"`
}

func (f Fault) Error() string {
	return fmt.Sprintf("%s: %s: %s", f.Actor, f.Code, f.Message)
}

// KasResponse a KAS SOAP response.
type KasResponse struct {
	Return *Item `xml:"return"`
}

// Item an item of the KAS SOAP response.
type Item struct {
	Text  string  `xml:",chardata" json:"text,omitempty"`
	Type  string  `xml:"type,attr" json:"type,omitempty"`
	Raw   string  `xml:"nil,attr" json:"raw,omitempty"`
	Key   *Item   `xml:"key" json:"key,omitempty"`
	Value *Item   `xml:"value" json:"value,omitempty"`
	Items []*Item `xml:"item" json:"item,omitempty"`
}
