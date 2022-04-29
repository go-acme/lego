package internal

import "fmt"

type Record struct {
	ID    string `json:"id,omitempty"`
	Slug  string `json:"slug,omitempty"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
	TTL   int    `json:"ttl,omitempty"`
}

// CreateRecordResponse represents a response from Vercel's API after making a DNS record.
type CreateRecordResponse struct {
	UID     string `json:"uid"`
	Updated int    `json:"updated,omitempty"`
}

type APIErrorResponse struct {
	Error *APIError `json:"error"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Code, a.Message)
}
