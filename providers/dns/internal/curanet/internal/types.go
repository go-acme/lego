package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`

	// TODO(ldez): handle additional properties when `json/v2` will land.
}

func (a *APIError) Error() string {
	var msg []string

	if a.Type != "" {
		msg = append(msg, "type: "+a.Type)
	}

	if a.Title != "" {
		msg = append(msg, "title: "+a.Title)
	}

	if a.Status != 0 {
		msg = append(msg, fmt.Sprintf("status: %d", a.Status))
	}

	if a.Detail != "" {
		msg = append(msg, "detail: "+a.Detail)
	}

	if a.Instance != "" {
		msg = append(msg, "instance: "+a.Instance)
	}

	return strings.Join(msg, ", ")
}

type Record struct {
	ID       int64  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Data     string `json:"data,omitempty"`
}
