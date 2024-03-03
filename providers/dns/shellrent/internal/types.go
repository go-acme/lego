package internal

import "fmt"

type Response[T any] struct {
	Base
	Data T `json:"data"`
}

type Base struct {
	Code    int    `json:"error"`
	Message string `json:"message"`
}

func (b Base) Error() string {
	return fmt.Sprintf("code %d: %s", b.Code, b.Message)
}

type ServiceDetails struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	DomainID int    `json:"domain_id,omitempty"`
}

type DomainDetails struct {
	ID              int    `json:"id"`
	DomainName      string `json:"domain_name"`
	DomainNameASCII string `json:"domain_name_ascii"`
}

type Record struct {
	ID   int    `json:"id,omitempty"`
	Kind string `json:"kind,omitempty"`
	Host string `json:"host,omitempty"`
	TTL  int    `json:"ttl,omitempty"` // It can be set to the following values (number of seconds): 3600, 14400, 28800, 57600, 86400
}
