package internal

import (
	"fmt"
	"time"
)

// APIResponse is the FENO response envelope.
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Code    string `json:"code,omitempty"`
}

// APIError is the error returned for every non-2xx response, and for 2xx responses whose envelope reports success=false.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
	RetryAfter time.Duration
}

func (a *APIError) Error() string {
	msg := fmt.Sprintf("%d: %s", a.StatusCode, a.Message)

	if a.Code != "" {
		msg += fmt.Sprintf(" (%s)", a.Code)
	}

	if a.RequestID != "" {
		msg += fmt.Sprintf(" [request %s]", a.RequestID)
	}

	if a.RetryAfter > 0 {
		msg += fmt.Sprintf(" retry after %s", a.RetryAfter)
	}

	return msg
}

// Domain is the minimal domain shape returned to an `acme:write` key.
type Domain struct {
	Domain          string `json:"domain"`
	Status          string `json:"status"`
	FenoNameservers bool   `json:"fenoNameservers"`
}

// Record is a DNS record.
type Record struct {
	ID    int    `json:"id,omitempty"`
	Type  string `json:"type,omitempty"`
	Name  string `json:"name"` // Relative to the zone, "" is the apex.
	Value string `json:"value,omitempty"`
	TTL   int    `json:"ttl,omitempty"`
}

// RecordList is the payload of the record listing.
type RecordList struct {
	DomainName string   `json:"domainName"`
	Records    []Record `json:"records"`
	Filtered   bool     `json:"filtered"`
}
