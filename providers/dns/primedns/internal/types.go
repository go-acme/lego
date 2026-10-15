package internal

import "fmt"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Code, a.Message)
}

type APIRequest struct {
	Domain string `json:"domain,omitempty"`
	Name   string `json:"name,omitempty"`
	Value  string `json:"value"`
}

type APIResponse[T any] struct {
	Success bool      `json:"success"`
	Data    T         `json:"data"`
	Error   *APIError `json:"error"`
}

type Record struct {
	ID    int    `json:"record_id"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

type DeleteResponse struct {
	Deleted int `json:"deleted"`
}
