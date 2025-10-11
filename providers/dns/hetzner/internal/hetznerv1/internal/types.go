package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	ErrorInfo ErrorInfo `json:"error"`
}

type ErrorInfo struct {
	Code    string       `json:"code,omitempty"`
	Message string       `json:"message,omitempty"`
	Details ErrorDetails `json:"details,omitempty"`
}

type ErrorDetails struct {
	Announcement string       `json:"announcement,omitempty"`
	Limits       []LimitError `json:"limits,omitempty"`
	Fields       []FieldError `json:"fields,omitempty"`
}

type FieldError struct {
	Name     string   `json:"name,omitempty"`
	Messages []string `json:"messages,omitempty"`
}

type LimitError struct {
	Name string `json:"name,omitempty"`
}

func (a *APIError) Error() string {
	msg := fmt.Sprintf("%s: %s", a.ErrorInfo.Code, a.ErrorInfo.Message)

	if a.ErrorInfo.Details.Announcement != "" {
		msg += fmt.Sprintf(": %s", a.ErrorInfo.Details.Announcement)
	}

	for _, limit := range a.ErrorInfo.Details.Limits {
		msg += fmt.Sprintf("limit: %s", limit.Name)
	}

	for _, field := range a.ErrorInfo.Details.Fields {
		msg += fmt.Sprintf("field: %s: %s", field.Name, strings.Join(field.Messages, ", "))
	}

	return msg
}

type RRSet struct {
	ID         string            `json:"id,omitempty"`
	Name       string            `json:"name,omitempty"`
	Type       string            `json:"type,omitempty"`
	TTL        int               `json:"ttl,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Protection *Protection       `json:"protection,omitempty"`
	Records    []Record          `json:"records,omitempty"`
	ZoneID     int               `json:"zone,omitempty"`
}

type Protection struct {
	Change bool `json:"change,omitempty"`
}

type Record struct {
	Value   string `json:"value,omitempty"`
	Comment string `json:"comment,omitempty"`
}

type ActionResponse struct {
	Action *Action `json:"action,omitempty"`
}

type Action struct {
	ID        int         `json:"id,omitempty"`
	Command   string      `json:"command,omitempty"`
	Status    string      `json:"status,omitempty"`
	Progress  int         `json:"progress,omitempty"`
	Resources []Resources `json:"resources,omitempty"`
	ErrorInfo *ErrorInfo  `json:"error,omitempty"`
}

type Resources struct {
	ID   int    `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}
