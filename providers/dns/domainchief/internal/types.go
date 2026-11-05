package internal

import (
	"fmt"
	"strings"
	"time"
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

type Challenge struct {
	Name  string `json:"name" url:"name"`
	Value string `json:"value" url:"value"`
}

type ChallengeRecord struct {
	ID        string    `json:"id"`
	Domain    string    `json:"domain"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	TTL       int       `json:"ttl"`
	CreatedAt time.Time `json:"created_at"`
}
