package internal

import (
	"encoding/json"
	"fmt"
)

type NotFoundError struct {
	APIError
}

type APIError struct {
	CorrelationID string          `json:"correlationId,omitempty"`
	ErrorCode     string          `json:"errorCode,omitempty"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
	ErrorDetails  json.RawMessage `json:"errorDetails,omitempty"`
}

func (a *APIError) Error() string {
	msg := fmt.Sprintf("%s: %s (%s)", a.ErrorCode, a.ErrorMessage, a.CorrelationID)

	if len(a.ErrorDetails) > 0 {
		msg += fmt.Sprintf(": %s", string(a.ErrorDetails))
	}

	return msg
}

type RRSet struct {
	Content  []string `json:"content,omitempty"`
	Name     string   `json:"name,omitempty"`
	Editable bool     `json:"editable,omitempty"`
	TTL      int      `json:"ttl,omitempty"`
	Type     string   `json:"type,omitempty"`
}
