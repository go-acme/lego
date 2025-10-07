package internal

import "encoding/json"

type APIResponse struct {
	Status   string          `json:"api-status,omitempty"`
	Response json.RawMessage `json:"api-response,omitempty"`
	Error    string          `json:"api-error,omitempty"`
}

type Domain struct {
	DomainName       string `json:"domain-name,omitempty"`
	RegistrationDate string `json:"registration-date,omitempty"`
	ExpirationDate   string `json:"expiration-date,omitempty"`
	Status           string `json:"status,omitempty"`
}

type Record struct {
	ID    int    `json:"line,omitempty" url:"-"`
	Type  string `json:"type,omitempty" url:"type,omitempty"`
	Name  string `json:"name,omitempty" url:"name,omitempty"`
	TTL   int    `json:"ttl,omitempty" url:"ttl,omitempty"`
	Value string `json:"value,omitempty" url:"value,omitempty"`
}

type DomainsResponse struct {
	Domains map[string]Domain `json:"domains,omitempty"`
}

type AddRecordResponse struct {
	Record *Record `json:"record,omitempty"`
}

type ListRecordsResponse struct {
	Records []Record `json:"records,omitempty"`
}

type DeleteRecordResponse struct {
	Deleted *DeletedRecordInfo `json:"deleted,omitempty"`
}

type DeletedRecordInfo struct {
	Count int   `json:"count,omitempty"`
	Lines []int `json:"lines,omitempty"`
}
