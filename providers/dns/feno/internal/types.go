package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
}

func NewAPIError(statusCode int, headers http.Header, res *APIResponse) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Code:       res.Code,
		Message:    res.Error,
		RequestID:  headers.Get("X-Request-Id"),
	}
}

func (a *APIError) Error() string {
	msg := fmt.Sprintf("%d: %s", a.StatusCode, a.Message)

	if a.Code != "" {
		msg += fmt.Sprintf(" (%s)", a.Code)
	}

	if a.RequestID != "" {
		msg += fmt.Sprintf(" [request %s]", a.RequestID)
	}

	return msg
}

type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
	Code    string          `json:"code,omitempty"`
}

type Domain struct {
	Domain string `json:"domain"`
	Status string `json:"status"`
	FenoNS bool   `json:"fenoNameservers"`
}

type Record struct {
	ID    int    `json:"id,omitempty"`
	Type  string `json:"type,omitempty"`
	Name  string `json:"name"` // Relative to the zone, "" is the apex.
	Value string `json:"value,omitempty"`
	TTL   int    `json:"ttl,omitempty"`
}

type RecordList struct {
	DomainName string   `json:"domainName"`
	Records    []Record `json:"records"`
	Filtered   bool     `json:"filtered"`
}
