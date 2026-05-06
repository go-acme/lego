package internal

import "strings"

type APIError struct {
	Errors []string `json:"error"`
}

func (a *APIError) Error() string {
	return strings.Join(a.Errors, ": ")
}

type Zones struct {
	ServiceIDs []string `json:"service_ids"`
	Zones      []Zone   `json:"zones"`
}

type Zone struct {
	DomainID  string `json:"domain_id"`
	Name      string `json:"name"`
	ServiceID string `json:"service_id"`
}

type Records struct {
	ServiceID int      `json:"service_id"`
	Name      string   `json:"name"`
	Records   []Record `json:"records"`
}

type Record struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Type     string `json:"type,omitempty"`
	Content  string `json:"content,omitempty"`
}
