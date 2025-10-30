package internal

import (
	"fmt"
	"strings"
)

// DNSRecord a DNS record.
type DNSRecord struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	Data string `json:"data"`
	TTL  int    `json:"ttl,omitempty"`

	Priority int    `json:"priority,omitempty"`
	Port     int    `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Service  string `json:"service,omitempty"`
	Weight   int    `json:"weight,omitempty"`
}

type APIError struct {
	Code    string  `json:"code,omitempty"`
	Fields  []Field `json:"fields,omitempty"`
	Message string  `json:"message,omitempty"`
}

func (a APIError) Error() string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("%s: %s", a.Code, a.Message))

	for _, field := range a.Fields {
		msg.WriteString(" ")
		msg.WriteString(field.String())
	}

	return msg.String()
}

type Field struct {
	Code        string `json:"code,omitempty"`
	Message     string `json:"message,omitempty"`
	Path        string `json:"path,omitempty"`
	PathRelated string `json:"pathRelated,omitempty"`
}

func (f Field) String() string {
	msg := fmt.Sprintf("[%s: %s", f.Code, f.Message)

	if f.Path != "" {
		msg += fmt.Sprintf(" (path=%s)", f.Path)
	}

	if f.PathRelated != "" {
		msg += fmt.Sprintf(" (pathRelated=%s)", f.PathRelated)
	}

	msg += "]"

	return msg
}
