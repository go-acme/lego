package internal

import (
	"fmt"
	"time"
)

// Zone represents a DNS zone in Gigahost.
type Zone struct {
	ID          string `json:"zone_id"`
	Name        string `json:"zone_name"`
	Type        string `json:"zone_type"`
	Active      string `json:"zone_active"`
	Status      string `json:"domain_status"`
	ExpiryDate  string `json:"domain_expiry_date"`
	RecordCount int    `json:"record_count"`
}

// Record represents a DNS record.
type Record struct {
	ID       string `json:"record_id"`
	Name     string `json:"record_name"`
	Type     string `json:"record_type"`
	Value    string `json:"record_value"`
	TTL      int    `json:"record_ttl"`
	Priority *int   `json:"record_priority,omitempty"`
}

// AuthResponse represents authentication response.
type AuthResponse struct {
	Meta Meta     `json:"meta"`
	Data AuthData `json:"data"`
}

// AuthData represents the data portion of auth response.
type AuthData struct {
	Token       string `json:"token"`
	TokenExpire int64  `json:"token_expire"`
	CustomerID  string `json:"customer_id"`
}

// Token represents a cached authentication token.
type Token struct {
	Token    string
	Deadline time.Time
}

// Meta represents API response metadata.
type Meta struct {
	Status        int    `json:"status"`
	StatusMessage string `json:"status_message"`
}

// ZonesResponse represents list zones response.
type ZonesResponse struct {
	Meta Meta   `json:"meta"`
	Data []Zone `json:"data"`
}

// RecordsResponse represents list records response.
type RecordsResponse struct {
	Meta Meta     `json:"meta"`
	Data []Record `json:"data"`
}

// CreateRecordRequest represents create record request.
type CreateRecordRequest struct {
	Name     string `json:"record_name,omitempty"`
	Type     string `json:"record_type,omitempty"`
	Value    string `json:"record_value"`
	TTL      int    `json:"record_ttl,omitempty"`
	Priority *int   `json:"record_priority,omitempty"`
}

// APIError represents an API error.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: %d - %s", e.StatusCode, e.Message)
}
