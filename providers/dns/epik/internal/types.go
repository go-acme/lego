package internal

import (
	"fmt"
	"strings"
)

type RecordRequest struct {
	Host string `json:"HOST,omitempty"`
	Type string `json:"TYPE,omitempty"`
	Data string `json:"DATA,omitempty"`
	Aux  int    `json:"AUX,omitempty"`
	TTL  int    `json:"TTL,omitempty"`
}

type SetHostRecords struct {
	Payload []RecordRequest `json:"set_host_records_payload"`
}

type CreateHostRecords struct {
	Payload RecordRequest `json:"create_host_records_payload"`
}

type Data struct {
	Code        int    `json:"code,omitempty"`
	Message     string `json:"message,omitempty"`
	Description string `json:"description,omitempty"`
}

type APIError struct {
	Errors []Data `json:"errors"`
}

func (a APIError) Error() string {
	var parts []string
	for _, data := range a.Errors {
		parts = append(parts, fmt.Sprintf("code: %d, message: %s, description: %s", data.Code, data.Message, data.Description))
	}

	return strings.Join(parts, ", ")
}

type Record struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
	AUX  int    `json:"aux"`
	TTL  int    `json:"ttl"`
}

type GetDNSRecordResponse struct {
	Data struct {
		Name    string   `json:"name"`
		Code    int      `json:"code"`
		Records []Record `json:"records"`
	} `json:"data"`
}
