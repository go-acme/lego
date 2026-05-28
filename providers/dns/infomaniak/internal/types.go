package internal

import (
	"fmt"
	"strings"
)

type APIResponse[T any] struct {
	Result string    `json:"result"`
	Data   T         `json:"data"`
	Error  *APIError `json:"error"`
}

type APIError struct {
	Code        string         `json:"code"`
	Description string         `json:"description"`
	Errors      []ErrorDetails `json:"errors"`
}

func (a *APIError) Error() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "[%s] %s", a.Code, a.Description)

	for _, err := range a.Errors {
		_, _ = fmt.Fprintf(msg, " (%s: %s)", err.Code, err.Description)
	}

	return msg.String()
}

type ErrorDetails struct {
	Code        string         `json:"code"`
	Description string         `json:"description"`
	Context     map[string]any `json:"context"`
}

type RecordRequest struct {
	Source string `json:"source,omitempty"`
	Target string `json:"target,omitempty"`
	TTL    int    `json:"ttl,omitempty"`
	Type   string `json:"type,omitempty"`
}

type Record struct {
	ID        int    `json:"id,omitempty"`
	Source    string `json:"source,omitempty"`
	SourceIDN string `json:"source_idn,omitempty"`
	Type      string `json:"type,omitempty"`
	TTL       int    `json:"ttl,omitempty"`
	Target    string `json:"target,omitempty"`
}

type DomainParams struct {
	AccountID int    `url:"account_id,omitempty"`
	OrderBy   string `url:"order_by,omitempty"`
	OrderDir  string `url:"order_dir,omitempty"`
	Search    string `url:"search,omitempty"`
	Page      int    `url:"page,omitempty"`
	PerPage   int    `url:"per_page,omitempty"`
}
