package internal

import (
	"fmt"
	"strings"
)

func ToStringSlice[T fmt.Stringer](values []T) []string {
	var s []string

	for _, value := range values {
		s = append(s, value.String())
	}

	return s
}

func Join[T fmt.Stringer](values []T, sep string) string {
	return strings.Join(ToStringSlice(values), sep)
}
