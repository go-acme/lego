package internal

import (
	"fmt"
	"strconv"
	"strings"
)

type APIError struct {
	Code           string   `json:"code"`
	Message        string   `json:"error"`
	Field          string   `json:"field"`
	Fields         []string `json:"fields"`
	RequiredScopes []string `json:"required_scopes"`
	RetryAfter     int      `json:"retry_after"`
}

func (a *APIError) Error() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "%s: %s", a.Code, a.Message)

	if a.Field != "" {
		msg.WriteString(": " + a.Field)
	}

	for _, field := range a.Fields {
		msg.WriteString(": " + field)
	}

	for _, scope := range a.RequiredScopes {
		msg.WriteString(": " + scope)
	}

	if a.RetryAfter > 0 {
		msg.WriteString(": retryAfter=" + strconv.Itoa(a.RetryAfter))
	}

	return msg.String()
}

type Record struct {
	ID       int64  `json:"id,omitempty"`
	Type     string `json:"type"`
	Hostname string `json:"hostname"`
	Text     string `json:"text,omitempty"`
	TTL      string `json:"ttl,omitempty"`
}
