package internal

import "fmt"

type APIError struct {
	RequestID string `json:"RequestId"`
	Message   string `json:"error"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s (%s)", a.Message, a.RequestID)
}

type Record struct {
	Domain string `url:"Domain,omitempty"`
	Host   string `url:"Host,omitempty"`
	Type   string `url:"Type,omitempty"`
	Value  string `url:"Value,omitempty"`
	Mx     string `url:"Mx,omitempty"`
	TTL    string `url:"Ttl,omitempty"`
}

type APIResponse struct {
	RequestID string `json:"RequestId"`
	ID        int    `json:"Id"`
}
