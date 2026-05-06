package internal

import (
	"fmt"
	"strings"
)

type apiResponse[T any] struct {
	Data T `json:"data,omitempty"`
}

type zonesData struct {
	Zones []Zone `json:"zones"`
}

type recordData struct {
	Record *Record `json:"record"`
}

type APIError struct {
	Message    string `json:"message,omitempty"`
	Code       string `json:"code,omitempty"`
	RequestID  string `json:"requestId,omitempty"`
	StatusCode int    `json:"-"`
}

func (a APIError) Error() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "%d: %s", a.StatusCode, a.Message)

	if a.Code != "" {
		_, _ = fmt.Fprintf(msg, " (%s)", a.Code)
	}

	if a.RequestID != "" {
		_, _ = fmt.Fprintf(msg, " [requestId=%s]", a.RequestID)
	}

	return msg.String()
}

type Zone struct {
	ServerID    string `json:"server_id,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
	DomainID    string `json:"domain_id,omitempty"`
	Domain      string `json:"domain,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}

type Record struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Type   string `json:"type,omitempty"`
	Value  string `json:"value,omitempty"`
	TTL    int    `json:"ttl,omitempty"`
	Status string `json:"status,omitempty"`
}
