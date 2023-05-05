package internal

import (
	"encoding/xml"
	"fmt"
)

// Record describes a DNS record returned by the Namecheap DNS gethosts API.
// Namecheap uses the term "host" to refer to all DNS records that include
// a host field (A, AAAA, CNAME, NS, TXT, URL).
type Record struct {
	Type    string `xml:",attr"`
	Name    string `xml:",attr"`
	Address string `xml:",attr"`
	MXPref  string `xml:",attr"`
	TTL     string `xml:",attr"`
}

// apiError describes an error record in a namecheap API response.
type apiError struct {
	Number      int    `xml:",attr"`
	Description string `xml:",innerxml"`
}

func (a apiError) Error() string {
	return fmt.Sprintf("%s [%d]", a.Description, a.Number)
}

type setHostsResponse struct {
	XMLName xml.Name   `xml:"ApiResponse"`
	Status  string     `xml:"Status,attr"`
	Errors  []apiError `xml:"Errors>Error"`
	Result  struct {
		IsSuccess string `xml:",attr"`
	} `xml:"CommandResponse>DomainDNSSetHostsResult"`
}

type getHostsResponse struct {
	XMLName xml.Name   `xml:"ApiResponse"`
	Status  string     `xml:"Status,attr"`
	Errors  []apiError `xml:"Errors>Error"`
	Hosts   []Record   `xml:"CommandResponse>DomainDNSGetHostsResult>host"`
}
