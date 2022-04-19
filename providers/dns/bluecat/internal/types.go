package internal

import "fmt"

// Entity JSON body for Bluecat entity requests.
type Entity struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Properties string `json:"properties"`
}

// EntityResponse JSON body for Bluecat entity responses.
type EntityResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Properties string `json:"properties"`
}

type APIError struct {
	StatusCode int
	Resource   string
	Message    string
}

func (a APIError) Error() string {
	return fmt.Sprintf("resource: %s, status code: %d, message: %s", a.Resource, a.StatusCode, a.Message)
}
