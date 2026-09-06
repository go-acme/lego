package internal

import "fmt"

type APIError struct {
	Message string `json:"error"`
	Code    string `json:"code"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s (%s)", a.Message, a.Code)
}

type Data struct {
	FQDN  string `json:"fqdn"`
	Value string `json:"value"`
}
