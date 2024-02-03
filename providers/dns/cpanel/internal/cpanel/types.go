package cpanel

import (
	"fmt"
	"strings"
)

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

func toError[T any](r Result[T]) error {
	return fmt.Errorf("error: %s: %s", strings.Join(r.Errors, ", "), strings.Join(r.Messages, ", "))
}
