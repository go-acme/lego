package internal

import (
	"encoding/json"
	"fmt"
	"strings"
)

const successResult = "success"

// APIResponse is the representation of an API response.
type APIResponse struct {
	Status string `json:"status"`

	Answer *Answer `json:"answer,omitempty"`

	ErrorCode string `json:"error_code,omitempty"`
	ErrorText string `json:"error_text,omitempty"`
}

func (a APIResponse) Error() string {
	return fmt.Sprintf("API %s: %s: %s", a.Status, a.ErrorCode, a.ErrorText)
}

// HasError returns an error is the response contains an error.
func (a APIResponse) HasError() error {
	if a.Status != successResult {
		return a
	}

	if a.Answer == nil || a.Status != successResult || a.Answer.Status != successResult {
		return a.Answer
	}

	return nil
}

// Answer is the representation of an API response answer.
type Answer struct {
	Status string          `json:"status,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`

	Errors    []AnswerError `json:"errors,omitempty"`
	ErrorCode string        `json:"error_code,omitempty"`
	ErrorText string        `json:"error_text,omitempty"`
}

type AnswerError struct {
	ErrorCode string `json:"error_code,omitempty"`
	ErrorText string `json:"error_text,omitempty"`
}

func (a Answer) Error() string {
	parts := []string{fmt.Sprintf("API answer %s", a.Status)}

	if a.ErrorCode != "" {
		parts = append(parts, a.ErrorCode)
	}

	if a.ErrorText != "" {
		parts = append(parts, a.ErrorText)
	}

	if len(a.Errors) > 0 {
		for _, e := range a.Errors {
			parts = append(parts, e.ErrorCode, e.ErrorText)
		}
	}

	return strings.Join(parts, ": ")
}

// ChangeRecordsRequest - data representation for data change request.
type ChangeRecordsRequest struct {
	Fqdn    string     `json:"fqdn,omitempty"`
	Records RecordList `json:"records,omitempty"`
}

// RecordList List of entries (in this case only described TXT).
type RecordList struct {
	TXT []Record `json:"TXT,omitempty"`
}

// Record record structure.
type Record struct {
	Priority int    `json:"priority,omitempty"`
	Value    string `json:"value,omitempty"`
}
