package internal

import "fmt"

type Zone struct {
	Name string `json:"name"`
}

type RRSet struct {
	TTL     int       `json:"ttl"`
	Records []Records `json:"resource_records"`
}

type Records struct {
	Content []string `json:"content"`
}

type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"error,omitempty"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("%d: %s", a.StatusCode, a.Message)
}
