package internal

import "fmt"

type Content struct {
	Text string `json:"text,omitempty"`
}

type Record struct {
	ID       string    `json:"id,omitempty"`
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	TTL      int       `json:"ttl"`
	Contents []Content `json:"contents"`
}

type RecordResponse struct {
	Status string `json:"status"`
	Data   Record `json:"data"`
}

type RecordsResponse struct {
	Status string   `json:"status"`
	Data   []Record `json:"data"`
}

type APIError struct {
	StatusCode   int    `json:"statusCode"`
	ErrorCode    string `json:"error"`
	ErrorMessage string `json:"message"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.ErrorCode, a.ErrorMessage)
}
