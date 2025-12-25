package internal

import (
	"encoding/json"
	"strings"
)

type APIResponse[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type ErrorInfo struct {
	ErrorCode    string          `json:"error_code"`
	ErrorDetails json.RawMessage `json:"error_details"`
}

type APIError APIResponse[ErrorInfo]

func (a *APIError) Error() string {
	var msg strings.Builder

	msg.WriteString(a.Status)
	msg.WriteString(": ")
	msg.WriteString(a.Message)
	msg.WriteString(" ")
	msg.WriteString(a.Data.ErrorCode)
	msg.WriteString(": ")
	msg.Write(a.Data.ErrorDetails)

	return msg.String()
}

type Record struct {
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"`
	Content string `json:"content,omitempty"`
	TTL     string `json:"ttl,omitempty"`
}
