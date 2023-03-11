package internal

import "fmt"

// APIException defines model for apiException.
type APIException struct {
	Message    string `json:"message,omitempty"`
	StatusCode int32  `json:"statusCode,omitempty"`
	Type       string `json:"type,omitempty"`
}

func (a APIException) Error() string {
	return fmt.Sprintf("%d: %s: %s", a.StatusCode, a.Type, a.Message)
}

// APIResponse defines model for apiResponse.
type APIResponse struct {
	Exception  *APIException `json:"exception,omitempty"`
	StatusCode int32         `json:"statusCode,omitempty"`
}

// DNSRecord defines model for dnsRecords.
type DNSRecord struct {
	ID         int64  `json:"id,omitempty"`
	Type       string `json:"recordType,omitempty"`
	DomainID   int64  `json:"domainId,omitempty"`
	DomainName string `json:"domainName,omitempty"`
	NodeName   string `json:"nodeName,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	State      bool   `json:"state,omitempty"`
	Content    string `json:"content,omitempty"`
	TextData   string `json:"textData,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
}

// DNSHostname defines model for DNS.hostname.
type DNSHostname struct {
	*APIException

	ID         int64  `json:"id,omitempty"`
	DomainName string `json:"domainName,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	Node       string `json:"node,omitempty"`
}

// RecordsResponse defines model for recordsResponse.
type RecordsResponse struct {
	*APIException

	DNSRecords []DNSRecord `json:"dnsRecords,omitempty"`
}

// RecordResponse defines model for recordResponse.
type RecordResponse struct {
	*APIException

	DNSRecord
}
