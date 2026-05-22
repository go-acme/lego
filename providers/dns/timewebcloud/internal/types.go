package internal

import "fmt"

type DNSRecordRequest struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type DNSRecord struct {
	ID   int    `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Fqdn string `json:"fqdn,omitempty"`
	Data Data   `json:"data"`
}

type Data struct {
	Value     string `json:"value,omitempty"`
	Subdomain string `json:"sub,omitempty"`
}

type CreateRecordResponse struct {
	DNSRecord *DNSRecord `json:"dns_record,omitempty"`
}

type ErrorResponse struct {
	StatusCode int    `json:"status_code,omitempty"`
	ErrorCode  string `json:"error_code,omitempty"`
	Message    string `json:"message,omitempty"`
	ResponseID string `json:"response_id,omitempty"`
}

func (a ErrorResponse) Error() string {
	return fmt.Sprintf("%d: %s (%s) [%s]", a.StatusCode, a.Message, a.ErrorCode, a.ResponseID)
}
