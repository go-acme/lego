package internal

import (
	"fmt"
)

type apiResponse[T any] struct {
	Data T `json:"data"`
}

type APIError struct {
	Message    string         `json:"message,omitempty"`
	Errors     map[string]any `json:"errors,omitempty"`
	StatusCode int            `json:"-"`
}

func (a APIError) Error() string {
	msg := fmt.Sprintf("%d: %s", a.StatusCode, a.Message)
	for k, v := range a.Errors {
		msg += fmt.Sprintf(" %s: %v", k, v)
	}
	return msg
}

type Zone struct {
	ID          int    `json:"id"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	TTL         int    `json:"ttl,omitempty"`
	Nameserver  string `json:"nameserver,omitempty"`
	Dnssec      bool   `json:"dnssec,omitempty"`
	DnssecEmail string `json:"dnssec_email,omitempty"`
}

type Record struct {
	ID      int    `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	Name    string `json:"name,omitempty"`
	Zone    string `json:"zone,omitempty"`
	Text    string `json:"text,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
	Comment string `json:"comment,omitempty"`
}
