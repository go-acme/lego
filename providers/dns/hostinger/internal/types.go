package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	Message       string              `json:"message,omitempty"`
	Errors        map[string][]string `json:"errors,omitempty"`
	CorrelationID string              `json:"correlation_id,omitempty"`
}

func (a *APIError) Error() string {
	msg := fmt.Sprintf("%s: %s", a.CorrelationID, a.Message)

	for field, values := range a.Errors {
		msg += fmt.Sprintf(": %s: %s", field, strings.Join(values, ", "))
	}

	return msg
}

type ZoneRequest struct {
	Overwrite bool        `json:"overwrite"`
	Zone      []RecordSet `json:"zone,omitempty"`
}

type RecordSet struct {
	Name    string   `json:"name,omitempty"`
	Records []Record `json:"records,omitempty"`
	TTL     int      `json:"ttl,omitempty"`
	Type    string   `json:"type,omitempty"`
}

type Record struct {
	Content    string `json:"content,omitempty"`
	IsDisabled bool   `json:"is_disabled,omitempty"`
}
