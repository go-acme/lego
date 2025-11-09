package internal

import "fmt"

type Zone struct {
	Name string `json:"name"`
}

type RRSet struct {
	TTL             int              `json:"ttl"`
	ResourceRecords []ResourceRecord `json:"resource_records"`
}

type ResourceRecord struct {
	Content []string `json:"content"`
	Enabled *bool    `json:"enabled,omitempty"`
}

type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"error,omitempty"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("%d: %s", a.StatusCode, a.Message)
}
