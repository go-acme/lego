package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	Type     string              `json:"type"`
	Title    string              `json:"title"`
	Status   int                 `json:"status"`
	Detail   string              `json:"detail"`
	Instance string              `json:"instance"`
	Errors   map[string][]string `json:"errors"`
}

func (a *APIError) Error() string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("%d: %s: %s: %s: %s", a.Status, a.Type, a.Title, a.Detail, a.Instance))

	for s, values := range a.Errors {
		msg.WriteString(fmt.Sprintf(": %s: %s", s, strings.Join(values, ", ")))
	}

	return msg.String()
}

type Record struct {
	ID       int64  `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Name     string `json:"name,omitempty"`
	Data     string `json:"data,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Port     int    `json:"port,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	Flags    int    `json:"flags,omitempty"`
	Tag      string `json:"tag,omitempty"`
}

type APIResponse struct {
	DomainRecord *Record `json:"domain_record"`
}
