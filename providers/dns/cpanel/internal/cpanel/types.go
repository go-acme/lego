package cpanel

import (
	"fmt"
	"strings"
)

type APIResponse[T any] struct {
	Metadata Metadata `json:"metadata"`
	Data     T        `json:"data,omitempty"`

	Status   int      `json:"status,omitempty"`
	Messages []string `json:"messages,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}

type Metadata struct {
	Transformed int `json:"transformed,omitempty"`
}

func toError[T any](r APIResponse[T]) error {
	return fmt.Errorf("error(%d): %s: %s", r.Status, strings.Join(r.Errors, ", "), strings.Join(r.Messages, ", "))
}
