package internal

import "fmt"

type APIError struct {
	Status            int    `json:"status,omitempty"`
	StatusDescription string `json:"status_description,omitempty"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("%d: %s", e.Status, e.StatusDescription)
}

type Response[T any] struct {
	Status            int    `json:"status,omitempty"`
	StatusDescription string `json:"status_description,omitempty"`
	Data              T      `json:"data,omitempty"`
}

type RecordData struct {
	Domain  string   `json:"domain,omitempty"`
	Records []Record `json:"records,omitempty"`
}

type Record struct {
	Type  string `json:"type,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
	TTL   int    `json:"ttl,omitempty"`
}

type DomainData struct {
	Domains []Domain `json:"domains"`
}

type Domain struct {
	ID          int      `json:"id"`
	Domain      string   `json:"domain"`
	RenewalDate string   `json:"renewal_date"`
	Status      string   `json:"status"`
	StatusID    int      `json:"status_id"`
	Tags        []string `json:"tags"`
}
