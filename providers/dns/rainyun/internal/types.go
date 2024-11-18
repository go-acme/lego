package internal

import "fmt"

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%d: %s", a.Code, a.Message)
}

type Record struct {
	ID       int    `json:"record_id,omitempty"`
	Host     string `json:"host,omitempty"`
	Priority int    `json:"level,omitempty"`
	Line     string `json:"line,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value,omitempty"`
}

type Domain struct {
	ID     int    `json:"record_id,omitempty"`
	Domain string `json:"domain,omitempty"`
}

type APIResponse[T any] struct {
	Code int      `json:"code"`
	Data *Data[T] `json:"data"`
}

// TODO(ldez) based on the API style, this structure is probably wrong.

type Data[T any] struct {
	TotalRecords int `json:"TotalRecords"` // TODO(ldez) based on the API style, the JSON name is probably wrong.
	Records      []T `json:"Records"`      // TODO(ldez) based on the API style, the JSON name is probably wrong.
}
