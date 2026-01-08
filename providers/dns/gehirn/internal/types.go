package internal

import "fmt"

type APIError struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%d: %s", a.Code, a.Description)
}

type Zone struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	DeletionProtection bool   `json:"deletion_protection"`
	CurrentVersionID   string `json:"current_version_id"`
}

type Record struct {
	ID      string      `json:"id,omitempty"`
	Name    string      `json:"name,omitempty"`
	Type    string      `json:"type,omitempty"`
	TTL     int         `json:"ttl,omitempty"`
	Records []RecordTXT `json:"records,omitempty"`
}

type RecordTXT struct {
	Data string `json:"data"`
}
