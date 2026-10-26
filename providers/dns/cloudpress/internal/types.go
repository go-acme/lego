package internal

import (
	"fmt"
	"strings"
)

type RecordType int

// https://docs.cloudpress.com/api/dns/#dns-records
const (
	RecordTypeA RecordType = iota
	RecordTypeAAAA
	RecordTypeCNAME
	RecordTypeTXT
	RecordTypeMX
	RecordTypeRedirect
	RecordTypeFlatten
	RecordTypePullZone
	RecordTypeSRV
	RecordTypeCAA
	RecordTypePTR
	RecordTypeScript
	RecordTypeNS
	RecordTypeSVCB
	RecordTypeHTTPS
	RecordTypeTLSA
)

type APIError struct {
	Code     string   `json:"code"`
	Messages []string `json:"errors"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Code, strings.Join(a.Messages, ", "))
}

type Record struct {
	ID      string     `json:"id,omitempty"`
	Type    RecordType `json:"record_type"`
	Name    string     `json:"name"`
	Value   string     `json:"value"`
	TTL     int        `json:"ttl,omitempty"`
	Comment string     `json:"comment,omitempty"`
}

type RecordEnvelope struct {
	Record *Record `json:"dns_record"`
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ZonesResponse struct {
	Zones []Zone `json:"dns_zones"`
}

type Pager struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
}
