package internal

import "fmt"

type APIResponse struct {
	TransactID string `json:"transactid"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	Code       int    `json:"code,omitempty"`
}

func (a APIResponse) Error() string {
	return fmt.Sprintf("%s(%d): %s (%s)", a.Status, a.Code, a.Message, a.TransactID)
}

type ListResponse struct {
	APIResponse

	TotalRecords int      `json:"total_records,omitempty"`
	Records      []Record `json:"records,omitempty"`
}

type Record struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
	TTL   int    `json:"ttl,omitempty"`
	Type  string `json:"type,omitempty"`
}

type RecordQuery struct {
	FullRecordName string `url:"fullrecordname"`
	Type           string `url:"type"`
	Value          string `url:"value,omitempty"`
	TTL            int    `url:"ttl,omitempty"`
}

type ListRecordQuery struct {
	Domain     string `url:"Domain"`
	FilterType string `url:"FilterType,omitempty"`
}
