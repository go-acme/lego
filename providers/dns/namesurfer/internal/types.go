package internal

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DNSNode represents a DNS record.
// http://95.128.3.201:8053/API/NSService_10#DNSNode
type DNSNode struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	Data string `json:"data,omitempty"`
	TTL  int    `json:"ttl,omitempty"`
}

// DNSZone represents a DNS zone.
// http://95.128.3.201:8053/API/NSService_10#DNSZone
type DNSZone struct {
	Name string `json:"name,omitempty"`
	View string `json:"view,omitempty"`
}

// APIRequest represents a JSON-RPC request.
type APIRequest struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
	ID     any    `json:"id"` // Can be int or string depending on API
}

// APIResponse represents a JSON-RPC response.
type APIResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *APIError       `json:"error"`
	ID     any             `json:"id"` // Can be int or string depending on API
}

// APIError represents a JSON-RPC error.
type APIError struct {
	Code       any      `json:"code"` // Can be int or string depending on API
	Filename   string   `json:"filename"`
	LineNumber int      `json:"lineno"`
	Message    string   `json:"string"`
	Detail     []string `json:"detail"`
}

func (e *APIError) Error() string {
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
