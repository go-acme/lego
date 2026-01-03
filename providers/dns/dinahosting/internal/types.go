package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	ResponseCode  int            `json:"responseCode"`
	Command       string         `json:"command"`
	TransactionID string         `json:"trId"`
	Errors        []ErrorDetails `json:"errors"`
}

func (a *APIError) Error() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "%d: %s (%s)", a.ResponseCode, a.Command, a.TransactionID)

	for _, details := range a.Errors {
		msg.WriteString(": " + details.String())
	}

	return msg.String()
}

type ErrorDetails struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Parameter string `json:"parameter"`
}

func (e *ErrorDetails) String() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "%d: %s", e.Code, e.Message)

	if e.Parameter != "" {
		_, _ = fmt.Fprintf(msg, " (%s)", e.Parameter)
	}

	return msg.String()
}

type TXTRecord struct {
	Domain   string `url:"domain,omitempty"`
	Hostname string `url:"hostname,omitempty"`
	Value    string `url:"value,omitempty"`
}
