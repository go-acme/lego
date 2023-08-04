package internal

import "fmt"

type Record struct {
	ID         int64  `json:"recordId,omitempty"`
	Address    string `json:"address,omitempty"`
	Exchange   string `json:"exchange,omitempty"`
	Flag       int64  `json:"flag,omitempty"`
	Name       string `json:"name,omitempty"`
	Port       int64  `json:"port,omitempty"`
	Preference int64  `json:"preference,omitempty"`
	Priority   int64  `json:"priority,omitempty"`
	Tag        string `json:"tag,omitempty"`
	Target     string `json:"target,omitempty"`
	Text       string `json:"text,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
	Type       string `json:"type,omitempty"`
	Value      string `json:"value,omitempty"`
	Weight     int64  `json:"weight,omitempty"`
}

type APIError struct {
	Code    int32    `json:"code"`
	Details []Detail `json:"details"`
	Message string   `json:"message"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("%d: %s: %v", a.Code, a.Message, a.Details)
}

type Detail struct {
	Type string `json:"@type"`
}

func (d Detail) String() string {
	return d.Type
}
