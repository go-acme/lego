package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	StatusCode    int      `json:"-"`
	ErrorMessage  string   `json:"error"`
	ErrorMessages []string `json:"errors"`
}

func (e *APIError) Error() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "%d: ", e.StatusCode)

	if e.ErrorMessage != "" {
		_, _ = fmt.Fprintf(msg, "%s", e.ErrorMessage)
	} else {
		_, _ = fmt.Fprintf(msg, "%s", strings.Join(e.ErrorMessages, "; "))
	}

	return msg.String()
}

type createRecordsRequest struct {
	Records []Record `json:"records"`
}

type Record struct {
	Host string `json:"host,omitempty"`
	TTL  int    `json:"ttl,omitempty"`
	Type string `json:"type,omitempty"`
	Data string `json:"data,omitempty"`
}

type recordResponse struct {
	Added   int    `json:"records_added"`
	Removed int    `json:"records_removed"`
	Message string `json:"message"`
}
