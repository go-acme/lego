package internal

import (
	"fmt"
	"time"
)

type Token struct {
	// The bearer token for use in API requests
	AccessToken string `json:"access_token"`
	TokenID     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	// Number in seconds before the expiration
	ExpiresIn       int    `json:"expires_in"`
	NotBeforePolicy int    `json:"not-before-policy"`
	Scope           string `json:"scope"`

	Deadline time.Time `json:"-"`
}

type authResponseError struct {
	ErrorMsg         string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (a authResponseError) Error() string {
	return fmt.Sprintf("%s: %s", a.ErrorMsg, a.ErrorDescription)
}

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
