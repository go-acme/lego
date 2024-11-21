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
	ID       int    `json:"record_id,omitempty" url:"record_id,omitempty"`
	Host     string `json:"host,omitempty" url:"host,omitempty"`
	Priority int    `json:"level,omitempty" url:"level,omitempty"`
	Line     string `json:"line,omitempty" url:"line,omitempty"`
	TTL      int    `json:"ttl,omitempty" url:"ttl,omitempty"`
	Type     string `json:"type,omitempty" url:"type,omitempty"`
	Value    string `json:"value,omitempty" url:"value,omitempty"`
}

type Domain struct {
	ID     int    `json:"id,omitempty"`
	Domain string `json:"domain,omitempty"`
}

type APIResponse[T any] struct {
	Code int      `json:"code"`
	Data *Data[T] `json:"data"`
}

type Data[T any] struct {
	TotalRecords int `json:"TotalRecords"`
	Records      []T `json:"Records"`
}
