package cpanel

import (
	"fmt"
	"strings"
)

type APIResponse[T any] struct {
	Metadata Metadata `json:"metadata,omitempty"`
	Messages []string `json:"messages,omitempty"`
	Status   int      `json:"status,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
	Data     T        `json:"data,omitempty"`
}

type Metadata struct {
	Transformed int `json:"transformed"`
}

func toError[T any](r APIResponse[T]) error {
	return fmt.Errorf("error(%d): %s: %s", r.Status, strings.Join(r.Errors, ", "), strings.Join(r.Messages, ", "))
}
