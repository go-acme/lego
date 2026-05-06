package zerossl

import (
	"fmt"
	"strings"
)

type APIResponse struct {
	Success bool `json:"success"`

	Kid     string `json:"eab_kid"`
	HmacKey string `json:"eab_hmac_key"`

	Error *ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code int    `json:"code"`
	Type string `json:"type"`
	Info string `json:"info"`
}

func (e *ErrorDetail) Error() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "%d: %s", e.Code, e.Type)

	if e.Info != "" {
		_, _ = fmt.Fprintf(msg, ": %s", e.Info)
	}

	return msg.String()
}
