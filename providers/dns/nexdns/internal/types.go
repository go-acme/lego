package internal

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

type APIResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

type APIError struct {
	Details APIErrorDetails `json:"error"`
}

func (a *APIError) Error() string {
	msg := fmt.Sprintf("%s: %s", a.Details.Code, a.Details.Message)

	if len(a.Details.Fields) == 0 {
		return msg
	}

	var fields []string

	for _, name := range slices.Sorted(maps.Keys(a.Details.Fields)) {
		fields = append(fields, fmt.Sprintf("%s: %s", name, strings.Join(a.Details.Fields[name], ", ")))
	}

	return fmt.Sprintf("%s (%s)", msg, strings.Join(fields, "; "))
}

type APIErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`

	Fields map[string][]string `json:"details,omitempty"`
}

type Zone struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type Record struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"`
	Content string `json:"content,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
}
