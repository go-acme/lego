package internal

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DNSNode represents a DNS record.
type DNSNode struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	Data string `json:"data,omitempty"`
	TTL  int    `json:"ttl,omitempty"`
}

// MinimalZone represents a DNS zone.
type MinimalZone struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	View string `json:"view,omitempty"`
}

// JSONRPCRequest represents a JSON-RPC 1.0 request.
type JSONRPCRequest struct {
	Method string `json:"method"`
	Params any    `json:"params"`
	ID     any    `json:"id"` // Can be int or string depending on API
}

// JSONRPCResponse represents a JSON-RPC response.
type JSONRPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *JSONRPCError   `json:"error"`
	ID     any             `json:"id"` // Can be int or string depending on API
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code       any      `json:"code"` // Can be int or string depending on API
	Filename   string   `json:"filename"`
	LineNumber int      `json:"lineno"`
	Message    string   `json:"string"`
	Detail     []string `json:"detail"`
}

func (e *JSONRPCError) Error() string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("code: %v", e.Code))

	if e.Filename != "" {
		msg.WriteString(fmt.Sprintf(", filename: %s", e.Filename))
	}

	if e.LineNumber > 0 {
		msg.WriteString(fmt.Sprintf(", line: %d", e.LineNumber))
	}

	if e.Message != "" {
		msg.WriteString(fmt.Sprintf(", message: %s", e.Message))
	}

	if len(e.Detail) > 0 {
		msg.WriteString(fmt.Sprintf(", detail: %v", strings.Join(e.Detail, " ")))
	}

	return msg.String()
}
