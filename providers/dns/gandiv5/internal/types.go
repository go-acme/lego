package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	Code    int    `json:"code"`
	Cause   string `json:"cause"`
	Message string `json:"message"`
	Object  string `json:"object"`
}

func (a *APIError) Error() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "(%d) %s", a.Code, a.Message)

	if a.Cause != "" {
		_, _ = fmt.Fprintf(msg, ": %s", a.Cause)
	}

	if a.Object != "" {
		_, _ = fmt.Fprintf(msg, ": %s", a.Object)
	}

	return msg.String()
}

// types for JSON responses with only a message.
type apiResponse struct {
	Message string `json:"message"`
	UUID    string `json:"uuid,omitempty"`
}

// Record TXT record representation.
type Record struct {
	TTL    int      `json:"rrset_ttl,omitempty"`
	Values []string `json:"rrset_values"`
	Name   string   `json:"rrset_name,omitempty"`
	Type   string   `json:"rrset_type,omitempty"`
}
