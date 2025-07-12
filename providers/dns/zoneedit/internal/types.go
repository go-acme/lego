package internal

import (
	"encoding/xml"
	"fmt"
)

type APIError struct {
	XMLName xml.Name `xml:"ERROR"`
	Text    string   `xml:",chardata"`
	Code    string   `xml:"CODE,attr"`
	Message string   `xml:"TEXT,attr"`
	Zone    string   `xml:"ZONE,attr"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("%s: %s (%s)", a.Code, a.Message, a.Zone)
}
