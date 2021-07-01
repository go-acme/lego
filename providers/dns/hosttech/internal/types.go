package internal

import (
	"encoding/json"
	"fmt"
)

type apiResponse struct {
	Data json.RawMessage `json:"data"`
}

type APIError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("%d: %s", a.StatusCode, a.Message)
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
