package internal

import (
	"fmt"
	"strings"
)

// APIError represents an error returned by the Zilore API.
// https://zilore.com/en/help/api/common/errors
type APIError struct {
	Status  string     `json:"status"`
	Info    *ErrorInfo `json:"error"`
	Details string     `json:"details"`
}

func (a *APIError) Error() string {
	var msg strings.Builder

	msg.WriteString(a.Status)

	if a.Details != "" {
		msg.WriteString(": " + a.Details)
	}

	if a.Info != nil {
		msg.WriteString(": " + a.Info.String())
	}

	return msg.String()
}

type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ErrorInfo) String() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

type Record struct {
	Name  string `url:"name"`
	Type  string `url:"type"`
	Value string `url:"value"`
	TTL   int    `url:"ttl"`
}

type APIResponse struct {
	Status   string          `json:"status"`
	Response *RecordResponse `json:"response"`
}

type RecordResponse struct {
	RecordID     int    `json:"record_id,omitempty"`
	RecordStatus string `json:"record_status,omitempty"`
}
