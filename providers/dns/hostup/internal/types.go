package internal

import (
	"fmt"
	"strings"
	"time"
)

// APIError is a generic error response.
// https://developer.hostup.se/errors/
type APIError struct {
	Type      string        `json:"type"`
	Title     string        `json:"title"`
	Status    int           `json:"status"`
	Detail    string        `json:"detail"`
	Code      string        `json:"code"`
	Instance  string        `json:"instance"`
	RequestID string        `json:"requestId"`
	Timestamp time.Time     `json:"timestamp"`
	Errors    []ErrorDetail `json:"errors"`
}

func (a *APIError) Error() string {
	var msg strings.Builder

	_, _ = fmt.Fprintf(&msg, "%d: %s %s %s",
		a.Status, a.Title, a.Detail, a.Code)

	_, _ = fmt.Fprintf(&msg, " (%s - %s)",
		a.Instance, a.RequestID)

	for _, detail := range a.Errors {
		_, _ = fmt.Fprintf(&msg, " %s", detail)
	}

	return msg.String()
}

type ErrorDetail struct {
	Pointer string `json:"pointer"`
	Detail  string `json:"detail"`
	Code    string `json:"code"`
}

func (e ErrorDetail) String() string {
	return fmt.Sprintf("[%s: %s %s]", e.Code, e.Detail, e.Pointer)
}

type APIResponse[T any] struct {
	Data       T      `json:"data"`
	HasMore    bool   `json:"hasMore,omitempty"`
	NextCursor string `json:"nextCursor,omitempty"`
}

type Zone struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ZoneStatus string `json:"zoneStatus"`
	IsDNSOnly  bool   `json:"isDnsOnly"`
	Kind       string `json:"kind"`
}

type Record struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	Port     int    `json:"port,omitempty"`
	Status   string `json:"status,omitempty"`
}
