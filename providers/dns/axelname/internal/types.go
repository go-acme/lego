package internal

import "fmt"

type APIError struct {
	ErrorCode string `json:"error_code,omitempty"`
	ErrorText string `json:"error_text,omitempty"`
	Result    string `json:"result,omitempty"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s: %s (%s)", a.Result, a.ErrorText, a.ErrorCode)
}

type APIResponse struct {
	APIError

	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type ListResponse struct {
	APIResponse

	Count int      `json:"count,omitempty"`
	List  []Record `json:"list,omitempty"`
}

type Record struct {
	ID    string `json:"id,omitempty" url:"id,omitempty"`
	Name  string `json:"name,omitempty" url:"name,omitempty"`
	Type  string `json:"type,omitempty" url:"type,omitempty"`
	Value string `json:"value,omitempty" url:"value,omitempty"`
	Prio  string `json:"prio,omitempty" url:"prio,omitempty"`
}
