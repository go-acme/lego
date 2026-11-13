package internal

import "fmt"

type APIError struct {
	Message string `json:"message"`
	Path    string `json:"path"`
	Status  int    `json:"status"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%d: %s: %s", a.Status, a.Path, a.Message)
}

type Record struct {
	Type     string `json:"recordType,omitempty"`
	Value    string `json:"value,omitempty"`
	Host     string `json:"host,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	Port     int    `json:"port,omitempty"`
}

type APIResponse struct {
	Status        string `json:"status"`
	Message       string `json:"message"`
	TransactionID string `json:"transactionId"`
}
