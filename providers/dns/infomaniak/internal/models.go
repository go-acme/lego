package internal

import (
	"encoding/json"
	"fmt"
)

// Record a DNS record.
type Record struct {
	ID     string `json:"id,omitempty"`
	Source string `json:"source,omitempty"`
	Type   string `json:"type,omitempty"`
	TTL    int    `json:"ttl,omitempty"`
	Target string `json:"target,omitempty"`
}

type DNSDomain struct {
	ID           uint64 `json:"id,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
}

type APIResponse struct {
	Result      string            `json:"result"`
	Data        json.RawMessage   `json:"data,omitempty"`
	ErrResponse *APIErrorResponse `json:"error,omitempty"`
}

type APIErrorResponse struct {
	Code        string             `json:"code"`
	Description string             `json:"description,omitempty"`
	Context     map[string]string  `json:"context,omitempty"`
	Errors      []APIErrorResponse `json:"errors,omitempty"`
}

func (a APIErrorResponse) Error() string {
	return fmt.Sprintf("code: %s, description: %s", a.Code, a.Description)
}
