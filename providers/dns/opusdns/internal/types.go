package internal

import "fmt"

type APIError struct {
	StatusCode int    `json:"-"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Detail     string `json:"detail"`
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("[status: %d] %s", e.StatusCode, e.Title)

	if e.Detail != "" {
		msg += ": " + e.Detail
	}

	return msg
}

type Record struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	TTL   int    `json:"ttl"`
	RData string `json:"rdata"`
}

type RecordOperation struct {
	Op     string `json:"op"`
	Record Record `json:"record"`
}

type PatchRequest struct {
	Ops []RecordOperation `json:"ops"`
}
