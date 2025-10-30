package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	StatusCode int      `json:"-"`
	Code       int      `json:"code"`
	Details    []string `json:"details"`
	Message    string   `json:"message"`
}

func (a *APIError) Error() string {
	var details string
	if len(a.Details) > 0 {
		details = " " + strings.Join(a.Details, ", ")
	}

	return fmt.Sprintf("code: %d, message: %s%s", a.Code, a.Message, details)
}

type APIRRSet struct {
	DNSZoneName string `json:"dns_zone_name,omitempty"`
	GroupName   string `json:"group_name,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	RecordName  string `json:"record_name,omitempty"`
	Type        string `json:"type,omitempty"`
	RRSet       RRSet  `json:"rrset"`
}

type RRSetRequest struct {
	DNSZoneName string `json:"dns_zone_name,omitempty"`
	GroupName   string `json:"group_name,omitempty"`
	RRSet       RRSet  `json:"rrset"`
}

type RRSet struct {
	Description string     `json:"description,omitempty"`
	TTL         int        `json:"ttl,omitempty"`
	TXTRecord   *TXTRecord `json:"txt_record,omitempty"`
}

type TXTRecord struct {
	Name   string   `json:"name,omitempty"`
	Values []string `json:"values,omitempty"`
}
