package internal

import (
	"time"
)

type APIResponse[T any] struct {
	Items []T `json:"items"`
}

type Zone struct {
	ID             string    `json:"id,omitempty"`
	ParentID       string    `json:"parent_id,omitempty"`
	Name           string    `json:"name,omitempty"`
	Valid          bool      `json:"valid,omitempty"`
	ValidationText string    `json:"validationText,omitempty"`
	Delegated      bool      `json:"delegated,omitempty"`
	LastCheck      time.Time `json:"lastCheck,omitzero"`
	CreatedAt      time.Time `json:"created_at,omitzero"`
	UpdatedAt      time.Time `json:"updated_at,omitzero"`
}

type Record struct {
	ZoneID  string   `json:"zone_id,omitempty"`
	Name    string   `json:"name,omitempty"`
	Type    string   `json:"type,omitempty"`
	Values  []string `json:"values,omitempty"`
	TTL     string   `json:"ttl,omitempty"`
	Enables bool     `json:"enables,omitempty"`
}
