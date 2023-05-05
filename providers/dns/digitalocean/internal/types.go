package internal

import "fmt"

// TxtRecordResponse represents a response from DO's API after making a TXT record.
type TxtRecordResponse struct {
	DomainRecord Record `json:"domain_record"`
}

type Record struct {
	ID   int    `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	Data string `json:"data,omitempty"`
	TTL  int    `json:"ttl,omitempty"`
}

type APIError struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.ID, a.Message)
}
