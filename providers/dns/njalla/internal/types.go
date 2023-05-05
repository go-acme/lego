package internal

import (
	"fmt"
)

// APIRequest represents an API request body.
type APIRequest struct {
	Method string `json:"method"`
	Params any    `json:"params"`
}

type Response interface {
	GetError() error
}

// APIResponse represents an API response body.
type APIResponse[T any] struct {
	ID     string    `json:"id"`
	RPC    string    `json:"jsonrpc"`
	Error  *APIError `json:"error,omitempty"`
	Result T         `json:"result,omitempty"`
}

func (a APIResponse[T]) GetError() error {
	if a.Error == (*APIError)(nil) {
		return nil
	}

	return a.Error
}

// APIError is an API error.
type APIError struct {
	Code    int
	Message string
}

func (a APIError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", a.Code, a.Message)
}

// Record is a DNS record.
type Record struct {
	ID      string `json:"id,omitempty"`
	Content string `json:"content,omitempty"`
	Domain  string `json:"domain,omitempty"`
	Name    string `json:"name,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
	Type    string `json:"type,omitempty"`
}

// Records is a list of DNS records.
type Records struct {
	Records []Record `json:"records,omitempty"`
}
