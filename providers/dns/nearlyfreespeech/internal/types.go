package internal

import "fmt"

type Record struct {
	Name string `url:"name,omitempty"`
	Type string `url:"type,omitempty"`
	Data string `url:"data,omitempty"`
	TTL  int    `url:"ttl,omitempty"`
}

type APIError struct {
	Message string `json:"error"`
	Debug   string `json:"debug"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Message, a.Debug)
}
