package internal

import "fmt"

type DomainInfoResponse struct {
	DomainInfo DomainInfo `json:"domainInfo"`
}

type DomainInfo struct {
	DNSRecords []Record `json:"dns_records"`
}

type Record struct {
	Type     string `json:"type,omitempty"`
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`
	Priority int    `json:"prio,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
}

type ErrorResponse struct {
	Message ErrorMessage `json:"error"`
}

type ErrorMessage struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e ErrorMessage) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}
