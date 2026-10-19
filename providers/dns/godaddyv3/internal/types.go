package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	Name            string         `json:"name,omitempty"`
	CorrelationID   string         `json:"correlationId,omitempty"`
	Message         string         `json:"message,omitempty"`
	InformationLink string         `json:"informationLink,omitempty"`
	Details         []ErrorDetails `json:"details,omitempty"`
}

type ErrorDetails struct {
	Field       string `json:"field,omitempty"`
	Value       string `json:"value,omitempty"`
	Location    string `json:"location,omitempty"`
	Issue       string `json:"issue,omitempty"`
	Description string `json:"description,omitempty"`
}

func (a *APIError) Error() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "%s: %s (%s)", a.Name, a.Message, a.CorrelationID)

	if a.InformationLink != "" {
		_, _ = fmt.Fprintf(msg, ", %s", a.InformationLink)
	}

	for _, detail := range a.Details {
		if detail.Field != "" {
			_, _ = fmt.Fprintf(msg, ", field=%s", detail.Field)
		}

		if detail.Value != "" {
			_, _ = fmt.Fprintf(msg, ", value=%s", detail.Value)
		}

		if detail.Location != "" {
			_, _ = fmt.Fprintf(msg, ", location=%s", detail.Location)
		}

		if detail.Issue != "" {
			_, _ = fmt.Fprintf(msg, ", issue=%s", detail.Issue)
		}

		if detail.Description != "" {
			_, _ = fmt.Fprintf(msg, ", description=%s", detail.Description)
		}
	}

	return msg.String()
}

type Record struct {
	ID       string `json:"recordId,omitempty"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Data     string `json:"data"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority,omitempty"`
	Service  string `json:"service,omitempty"`
	Port     int    `json:"port,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Flag     int    `json:"flag,omitempty"`
	Tag      string `json:"tag,omitempty"`
}
