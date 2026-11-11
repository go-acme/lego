package internal

import "fmt"

type ErrorEnvelope struct {
	Error *APIError `json:"error"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Code, a.Message)
}

type Challenge struct {
	Name  string `json:"name" url:"name"`
	Value string `json:"value" url:"value"`
}

type ResponseEnvelope[T any] struct {
	Data T `json:"data"`
}

type ChallengeRecord struct {
	ServiceID   int64    `json:"service_id"`
	Zone        string   `json:"zone"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	Values      int      `json:"values"`
	Changed     bool     `json:"changed"`
	Ready       bool     `json:"ready"`
	Nameservers []string `json:"nameservers"`
	TTL         int      `json:"ttl"`
}
