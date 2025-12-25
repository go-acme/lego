package internal

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type APIError struct{}

func (a *APIError) Error() string {
	// TODO implement me
	panic("implement me")
}

type Record struct {
	ID    int64  `json:"id,omitempty"`
	Host  string `json:"host,omitempty"`
	Type  string `json:"type,omitempty"`
	TTL   int    `json:"ttl,omitempty"`
	Value string `json:"value,omitempty"`
}

func CheckTTLValue(value int) error {
	enum := []int{120, 300, 600, 900, 1800, 3600, 7200, 18000, 43200, 86400}

	if slices.Contains(enum, value) {
		return nil
	}

	var values []string
	for _, e := range enum {
		values = append(values, strconv.Itoa(e))
	}

	return fmt.Errorf("invalid TTL value: %d, the possible values are: %s", value, strings.Join(values, ", "))
}
