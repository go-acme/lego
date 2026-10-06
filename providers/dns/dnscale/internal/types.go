package internal

import "fmt"

type APIError struct {
	Status  string          `json:"status"`
	Details APIErrorDetails `json:"error"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s: %s (%s)", a.Status, a.Details.Message, a.Details.Code)
}

type APIErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

type PaginatedZones struct {
	Zones      []Zone     `json:"zones"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Total   int  `json:"total"`
	Offset  int  `json:"offset"`
	Limit   int  `json:"limit"`
	Count   int  `json:"count"`
	HasMore bool `json:"has_more"`
}

type Pager struct {
	Limit  int `url:"limit"`
	Offset int `url:"offset"`
}

type RecordData struct {
	Message string  `json:"message"`
	Record  *Record `json:"record"`
}

type Record struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Content  string `json:"content,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

type Zone struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	RecordCount   int    `json:"record_count"`
	DNSSecEnabled bool   `json:"dnssec_enabled"`
}
