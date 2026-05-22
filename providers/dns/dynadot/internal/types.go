package internal

import "fmt"

type APIResponse struct {
	Code        int          `json:"code"`
	Message     string       `json:"message"`
	ErrorDetail *ErrorDetail `json:"error,omitempty"`
}

func (e *APIResponse) Error() string {
	msg := fmt.Sprintf("[code: %d] %s", e.Code, e.Message)

	if e.ErrorDetail != nil && e.ErrorDetail.Description != "" {
		msg += ": " + e.ErrorDetail.Description
	}

	return msg
}

type ErrorDetail struct {
	Description string `json:"description"`
}

type Record struct {
	Type   string `json:"record_type"`
	Value1 string `json:"record_value1"`
	Value2 string `json:"record_value2,omitempty"`
}

type SubRecord struct {
	Record

	SubHost string `json:"sub_host"`
}

type SetDNSRequest struct {
	Mains           []Record    `json:"dns_main_list,omitempty"`
	Subs            []SubRecord `json:"sub_list,omitempty"`
	TTL             int         `json:"ttl,omitempty"`
	AddDNSToCurrent bool        `json:"add_dns_to_current_setting,omitempty"`
}

type RemoveDNSRequest struct {
	Mains []Record    `json:"dns_main_list,omitempty"`
	Subs  []SubRecord `json:"sub_list,omitempty"`
}
