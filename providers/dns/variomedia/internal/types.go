package internal

import (
	"fmt"
	"strings"
)

type CreateDNSRecordRequest struct {
	Data Data `json:"data"`
}

type Data struct {
	Type       string    `json:"type"`
	Attributes DNSRecord `json:"attributes"`
}

type DNSRecord struct {
	RecordType string `json:"record_type,omitempty"`
	Name       string `json:"name,omitempty"`
	Domain     string `json:"domain,omitempty"`
	Data       string `json:"data,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
}

type APIError struct {
	Errors []ErrorItem `json:"errors"`
}

func (a APIError) Error() string {
	var parts []string
	for _, data := range a.Errors {
		parts = append(parts, fmt.Sprintf("status: %s, title: %s, id: %s", data.Status, data.Title, data.ID))
	}

	return strings.Join(parts, ", ")
}

type ErrorItem struct {
	Status string `json:"status,omitempty"`
	Title  string `json:"title,omitempty"`
	ID     string `json:"id,omitempty"`
}

type CreateDNSRecordResponse struct {
	Data struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			Status string `json:"status"`
		} `json:"attributes"`
		Links struct {
			QueueJob  string `json:"queue-job"`
			DNSRecord string `json:"dns-record"`
		} `json:"links"`
	} `json:"data"`
}

type GetJobResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			JobType string `json:"job_type"`
			Status  string `json:"status"`
		} `json:"attributes"`
		Links struct {
			Self   string `json:"self"`
			Object string `json:"object"`
		} `json:"links"`
	} `json:"data"`
}

type DeleteRecordResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			JobType string `json:"job_type"`
			Status  string `json:"status"`
		} `json:"attributes"`
		Links struct {
			Self   string `json:"self"`
			Object string `json:"object"`
		} `json:"links"`
	} `json:"data"`
}
