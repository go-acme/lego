package internal

import "fmt"

type Record struct {
	ID         string `json:"id"`
	Type       string `json:"type,omitempty"`
	RecordName string `json:"record_name,omitempty"`
	Content    string `json:"content,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
	Priority   int    `json:"priority,omitempty"`
	Service    string `json:"service,omitempty"`
	Weight     int    `json:"weight,omitempty"`
	Target     string `json:"target,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	Port       int    `json:"port,omitempty"`
}

type GetRecordsRequest struct {
	Skip       int    `url:"skip,omitempty"`
	Take       int    `url:"take,omitempty"`
	Type       string `url:"type,omitempty"`
	RecordName string `url:"record_name,omitempty"`
	Service    string `url:"service,omitempty"`
}

type APIError struct {
	ErrorCode string `json:"error_code"`
	ErrorText string `json:"error_text"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.ErrorCode, a.ErrorText)
}
