package internal

import "fmt"

// Record represents a DNS record.
type Record struct {
	Name  string `url:"name,omitempty"`
	Type  string `url:"type,omitempty"`
	Value string `url:"value,omitempty"`
	TTL   int    `url:"ttl,omitempty"`
}

// APIError represents a API error.
type APIError struct {
	Message string `json:"error,omitempty"`
	Result  string `json:"result,omitempty"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Message, a.Result)
}
