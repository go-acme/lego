package internal

import (
	"fmt"
	"strconv"
)

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
	ID          IntOrString `json:"id,omitempty"`
	Type        string      `json:"type,omitempty"`
	Host        string      `json:"host,omitempty"`
	TTL         int         `json:"ttl,omitempty"` // It can be set to the following values (number of seconds): 3600, 14400, 28800, 57600, 86400
	Destination string      `json:"destination,omitempty"`
}

type IntOrString int

func (m *IntOrString) Value() int {
	if m == nil {
		return 0
	}

	return int(*m)
}

func (m *IntOrString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	raw := string(data)
	if data[0] == '"' {
		var err error

		raw, err = strconv.Unquote(string(data))
		if err != nil {
			return err
		}
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return err
	}

	*m = IntOrString(v)

	return nil
}
