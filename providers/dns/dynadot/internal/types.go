package internal

import "fmt"

// MainRecord represents a main (apex) DNS record.
type MainRecord struct {
	RecordType   string `json:"record_type"`
	RecordValue1 string `json:"record_value1"`
	RecordValue2 string `json:"record_value2,omitempty"`
}

// SubRecord represents a sub-host DNS record.
type SubRecord struct {
	SubHost      string `json:"sub_host"`
	RecordType   string `json:"record_type"`
	RecordValue1 string `json:"record_value1"`
	RecordValue2 string `json:"record_value2,omitempty"`
}

// SetDNSRequest is the body for the set_dns endpoint.
// https://www.dynadot.com/domain/api-document
type SetDNSRequest struct {
	DNSMainList            []MainRecord `json:"dns_main_list,omitempty"`
	SubList                []SubRecord  `json:"sub_list,omitempty"`
	TTL                    int64        `json:"ttl,omitempty"`
	AddDNSToCurrentSetting bool         `json:"add_dns_to_current_setting,omitempty"`
}

// RemoveDNSRequest is the body for the remove_dns endpoint.
// https://www.dynadot.com/domain/api-document
type RemoveDNSRequest struct {
	DNSMainList []MainRecord `json:"dns_main_list,omitempty"`
	SubList     []SubRecord  `json:"sub_list,omitempty"`
}

// APIResponse mirrors the Dynadot RESTful v2 envelope.
//
// Unlike many APIs, Dynadot uses HTTP-style status codes inside the body and
// returns 200 on success.
type APIResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Error   *APIErrorObj `json:"error,omitempty"`
}

// APIErrorObj is the optional nested error object returned on failures.
type APIErrorObj struct {
	Description string `json:"description"`
}

// APIError is the error returned by the Dynadot API.
type APIError struct {
	Code        int
	Message     string
	Description string
}

func (e *APIError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("[code: %d] %s: %s", e.Code, e.Message, e.Description)
	}

	return fmt.Sprintf("[code: %d] %s", e.Code, e.Message)
}
