package internal

import "encoding/json"

type Record struct {
	DomainName string `json:"domainName,omitempty"`
	RecordName string `json:"recordName,omitempty"`
	Type       string `json:"type,omitempty"`
	Value      string `json:"value,omitempty"`
	Line       string `json:"line,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
	Status     int    `json:"status,omitempty"`
}

type APIResponse struct {
	Code      string          `json:"code"`
	Message   string          `json:"msg"`
	Data      json.RawMessage `json:"data"`
	RequestID string          `json:"requestId"`
}
