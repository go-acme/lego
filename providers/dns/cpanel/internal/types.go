package internal

import (
	"fmt"
	"strings"
)

type Record struct {
	DName      string   `json:"dname,omitempty"`
	TTL        int      `json:"ttl,omitempty"`
	RecordType string   `json:"record_type,omitempty"`
	Data       []string `json:"data,omitempty"`
	LineIndex  int      `json:"line_index,omitempty"`
}

type APIResponse[T any] struct {
	APIVersion int       `json:"apiversion"`
	Func       string    `json:"func"`
	Module     string    `json:"module"`
	Result     Result[T] `json:"result"`
}

type Result[T any] struct {
	Data     T        `json:"data"`
	Errors   []string `json:"errors"`
	Messages []string `json:"messages"`
	Metadata Metadata `json:"metadata"`
	Status   int      `json:"status"`
	Warnings []string `json:"warnings"`
}

type Metadata struct {
	Transformed int `json:"transformed"`
}

type ZoneSerial struct {
	NewSerial int `json:"new_serial"`
}

type ZoneRecord struct {
	LineIndex  int      `json:"line_index,omitempty"`
	Type       string   `json:"type,omitempty"`
	DataB64    []string `json:"data_b64,omitempty"`
	DNameB64   string   `json:"dname_b64,omitempty"`
	RecordType string   `json:"record_type,omitempty"`
	TTL        int      `json:"ttl,omitempty"`
}

func toError[T any](r Result[T]) error {
	return fmt.Errorf("error: %s: %s", strings.Join(r.Errors, ", "), strings.Join(r.Messages, ", "))
}
