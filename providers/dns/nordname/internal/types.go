package internal

import "fmt"

type APIError struct {
	Status    int    `json:"status"`
	Detail    string `json:"detail"`
	ErrorCode string `json:"error_code"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%d(%s): %s", a.Status, a.ErrorCode, a.Detail)
}

type Record struct {
	ID      string `json:"id,omitempty" url:"id,omitempty"`
	Name    string `json:"name" url:"name"`
	Type    string `json:"type" url:"type"`
	TTL     int    `json:"ttl" url:"ttl"`
	Content string `json:"content" url:"content"`
}
