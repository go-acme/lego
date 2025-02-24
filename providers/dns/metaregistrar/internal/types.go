package internal

import "fmt"

type APIError struct {
	Err          string `json:"error,omitempty"`
	Message      string `json:"message,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

func (e APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Err, e.Message)
	}

	return fmt.Sprintf("%s: %s", e.Err, e.ErrorMessage)
}

type Content struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Content  string `json:"content,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

type DnszoneUpdateRequest struct {
	Add    []Content `json:"add,omitempty"`
	Remove []Content `json:"rem,omitempty"`
}
