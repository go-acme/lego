package internal

import (
	"fmt"
	"strings"
	"time"
)

// Record status.
// https://docs.chief.tools/api/domainchief/~schemas#acmechallengestatus
const (
	StatusPending    = "pending"
	StatusPublishing = "publishing"
	StatusPublished  = "published"
	StatusFailed     = "failed"
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

// ChallengeRecord represents an ACME record.
// https://docs.chief.tools/api/domainchief/~schemas#acmechallenge
type ChallengeRecord struct {
	ID        string    `json:"id"`
	Domain    string    `json:"domain"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	TTL       int       `json:"ttl"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
