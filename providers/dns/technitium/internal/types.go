package internal

import "fmt"

type APIResponse[T any] struct {
	Status string `json:"status"` // ok/error/invalid-token

	Response T `json:"response"`

	ErrorMessage      string `json:"errorMessage"`
	StackTrace        string `json:"stackTrace"`
	InnerErrorMessage string `json:"innerErrorMessage"`
}

func (a *APIResponse[T]) Error() string {
	msg := fmt.Sprintf("Status: %s", a.Status)

	if a.ErrorMessage != "" {
		msg += fmt.Sprintf(", ErrorMessage: %s", a.ErrorMessage)
	}

	if a.StackTrace != "" {
		msg += fmt.Sprintf(", StackTrace: %s", a.StackTrace)
	}

	if a.InnerErrorMessage != "" {
		msg += fmt.Sprintf(", InnerErrorMessage: %s", a.InnerErrorMessage)
	}

	return msg
}

type AddRecordResponse struct {
	Zone        *Zone   `json:"zone"`
	AddedRecord *Record `json:"addedRecord"`
}

type Record struct {
	Name   string `json:"name,omitempty" url:"-"`
	Domain string `json:"domain,omitempty" url:"domain"`
	Type   string `json:"type,omitempty" url:"type"`
	Text   string `json:"text,omitempty" url:"text"`
}

type Zone struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
