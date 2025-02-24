package internal

import (
	"strings"
)

// APIError It's a mix of documented and undocumented fields.
// Note: the documentation is inconsistent: the names of property are not the same as the JSON sample.
// https://metaregistrar.dev/docu/metaapi/requests/response_ErrorResponse.html
type APIError struct {
	ResponseID   string `json:"responseId,omitempty"`
	Status       string `json:"status,omitempty"`
	Message      string `json:"message,omitempty"`
	Err          string `json:"error,omitempty"`
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

func (e *APIError) Error() string {
	var msg []string

	if e.Status != "" {
		msg = append(msg, e.Status)
	}

	if e.Err != "" {
		msg = append(msg, e.Err)
	}

	if e.ErrorCode != "" {
		msg = append(msg, e.ErrorCode)
	}

	if e.Message != "" {
		msg = append(msg, e.Message)
	}

	if e.ErrorMessage != "" {
		msg = append(msg, e.ErrorMessage)
	}

	return strings.Join(msg, ": ")
}

type Record struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Content  string `json:"content,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

// DNSZoneUpdateRequest is the representation of DnszoneUpdateRequest object.
// https://metaregistrar.dev/docu/metaapi/requests/request_DnszoneUpdateRequest.html
type DNSZoneUpdateRequest struct {
	Add    []Record `json:"add,omitempty"`
	Remove []Record `json:"rem,omitempty"`
}

// DNSZoneUpdateResponse is the representation of DnszoneUpdateResponse object.
// https://metaregistrar.dev/docu/metaapi/requests/response_DnszoneUpdateResponse.html
type DNSZoneUpdateResponse struct {
	ResponseID string `json:"responseId,omitempty"`
	Status     string `json:"status,omitempty"`
	Message    string `json:"message,omitempty"`
}
