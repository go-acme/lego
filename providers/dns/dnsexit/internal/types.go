package internal

import (
	"fmt"
	"strings"
)

type Record struct {
	Type      string `json:"type,omitempty"`
	Name      string `json:"name,omitempty"`
	Content   string `json:"content,omitempty"`
	TTL       int    `json:"ttl,omitempty"` // NOTE: ttl value is in minutes.
	Overwrite bool   `json:"overwrite,omitempty"`
}

type APIRequest struct {
	Domain string   `json:"domain,omitempty"`
	Add    []Record `json:"add,omitempty"`
	Delete []Record `json:"delete,omitempty"`
	Update []Record `json:"update,omitempty"`
}

// https://dnsexit.com/dns/dns-api/#server-reply

type APIResponse struct {
	Code    int      `json:"code,omitempty"`
	Details []string `json:"details,omitempty"`
	Message string   `json:"message,omitempty"`
}

func (a APIResponse) Error() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "%s (code=%d)", a.Message, a.Code)

	for _, detail := range a.Details {
		_, _ = fmt.Fprintf(msg, ", %s", detail)
	}

	return msg.String()
}
