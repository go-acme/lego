package internal

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	ErrorMessage string          `json:"errorMessage,omitempty"`
	ErrorDetails string          `json:"errorDetails,omitempty"`
	LogID        int             `json:"logID,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("message: %s, details: %s, logiD: %d, result: %s", a.ErrorMessage, a.ErrorDetails, a.LogID, a.Result)
}

type APIResponse[T any] struct {
	Result T   `json:"result,omitempty"`
	LogID  int `json:"logID,omitempty"`
}

type DNSInfo struct {
	DomainID      int            `json:"domainID,omitempty"`
	DNSRecordSets []DNSRecordSet `json:"dnsRecordSets,omitempty"`
}

type DNSRecordSet struct {
	Hostname string   `json:"hostname"`
	Type     string   `json:"type"`
	Records  []string `json:"records"`
}
