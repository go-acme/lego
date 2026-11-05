package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (a *APIError) Error() string {
	msg := new(strings.Builder)

	msg.WriteString(a.Message)

	if a.Code != "" {
		_, _ = fmt.Fprintf(msg, " (%s)", a.Code)
	}

	return msg.String()
}

type Record struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Name     string `json:"name,omitempty"`
	Content  string `json:"content,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"prio,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

type APIEnvelope[T any] struct {
	Data T `json:"data"`
}
