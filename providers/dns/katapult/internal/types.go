package internal

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	Code        string          `json:"code"`
	Description string          `json:"description"`
	Detail      json.RawMessage `json:"detail"`
}

func (a *APIError) Error() string {
	msg := fmt.Sprintf("%s: %s", a.Code, a.Description)

	if len(a.Detail) > 0 {
		msg += ": " + string(a.Detail)
	}

	return msg
}

type CreateRecordRequest struct {
	DNSZone    *DNSZone          `json:"dns_zone,omitempty"`
	Properties *RecordProperties `json:"properties,omitempty"`
}

type DNSZone struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type RecordProperties struct {
	Name     string         `json:"name,omitempty"`
	Type     string         `json:"type,omitempty"`
	TTL      int            `json:"ttl,omitempty"`
	Priority int            `json:"priority,omitempty"`
	Content  *RecordContent `json:"content,omitempty"`
}

type CreateRecordResponse struct {
	DNSRecord *DNSRecordResponse `json:"dns_record"`
}

type DNSRecordResponse struct {
	ID                string         `json:"id,omitempty"`
	Name              string         `json:"name,omitempty"`
	FullName          string         `json:"full_name,omitempty"`
	TTL               int            `json:"ttl,omitempty"`
	Type              string         `json:"type,omitempty"`
	Priority          int            `json:"priority,omitempty"`
	Content           string         `json:"content,omitempty"`
	ContentAttributes *RecordContent `json:"content_attributes,omitempty"`
}

type RecordContent struct {
	TXT *RecordTXT `json:"TXT,omitempty"`
}

type RecordTXT struct {
	Content string `json:"content,omitempty"`
}
