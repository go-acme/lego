package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	Detail string `json:"detail"`
	Data   []struct {
		Field   string `json:"field"`
		Details string `json:"details"`
	} `json:"data"`
}

func (a *APIError) Error() string {
	msg := []string{a.Detail}

	for _, datum := range a.Data {
		msg = append(msg, fmt.Sprintf("%s: %s", datum.Field, datum.Details))
	}

	return strings.Join(msg, ", ")
}

type Foo struct {
	Force bool     `json:"force,omitempty"`
	Items []Record `json:"items,omitempty"`
}

type Record struct {
	Type       string `json:"type,omitempty"`
	Name       string `json:"name,omitempty"`
	Value      string `json:"value,omitempty"`
	Address    string `json:"address,omitempty"`
	Nameserver string `json:"nameserver,omitempty"`
	AliasName  string `json:"aliasName,omitempty"`
	Pointer    string `json:"pointer,omitempty"`
	CName      string `json:"cname,omitempty"`
	Exchange   string `json:"exchange,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
}

type GetRecordsResponse struct {
	Items []Record `json:"items"`
	Total int      `json:"total"`
}
