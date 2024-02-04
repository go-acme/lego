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

// ---

type APIResponseOLD[T any] struct {
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

// ---

type Metadata struct {
	Transformed int `json:"transformed"`
}

func toErrorOLD[T any](r Result[T]) error {
	return fmt.Errorf("error(%d): %s: %s", r.Status, strings.Join(r.Errors, ", "), strings.Join(r.Messages, ", "))
}

func toError[T any](r APIResponse[T]) error {
	return fmt.Errorf("error(%d): %s: %s", r.Status, strings.Join(r.Errors, ", "), strings.Join(r.Messages, ", "))
}
