package internal

import "fmt"

type APIError struct {
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Code, a.Reason)
}

type Record struct {
	ID        string `json:"id,omitempty"`
	AccountID string `json:"account_id,omitempty"`
	DomainID  string `json:"domain_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Value     string `json:"value,omitempty"`
	Type      string `json:"type,omitempty"`
	TTL       int    `json:"ttl,omitempty"`
}

type Domain struct {
	ID        string `json:"id,omitempty"`
	AccountID string `json:"account_id,omitempty"`
	Name      string `json:"name,omitempty"`
}
