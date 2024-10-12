package internal

import "fmt"

type CreateRecordPayload struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	SubDomain string `json:"subdomain"`
}

type DNSRecord struct {
	Type string `json:"type"`
	ID   int    `json:"id,omitempty"`
}

type CreateRecordResponse struct {
	DNSRecord DNSRecord `json:"dns_record"`
}

type ErrorResponse struct {
	StatusCode int      `json:"status_code"`
	Message    []string `json:"message,omitempty"`
}

func (a ErrorResponse) Error() string {
	return fmt.Sprintf("%d: %s", a.StatusCode, a.Message)
}
