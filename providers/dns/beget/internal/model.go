package internal

import "fmt"

const successResult = "success"

// APIResponse is the representation of an API response.
type APIResponse struct {
	Status string `json:"status"`

	Answer *Answer `json:"answer,omitempty"`

	ErrorCode string `json:"error_code,omitempty"`
	ErrorText string `json:"error_text,omitempty"`
}

func (a APIResponse) Error() string {
	return fmt.Sprintf("API %s: %s: %s", a.Status, a.ErrorCode, a.ErrorText)
}

func (a Answer) Error() string {
	return fmt.Sprintf("API answer %s: %s: %s", a.Status, a.ErrorCode, a.ErrorText)
}

// HasError returns an error is the response contains an error.
func (a APIResponse) HasError() error {
	if a.Status != successResult {
		return a
	}

	if a.Answer == nil || a.Status != successResult || !a.Answer.Result {
		return a.Answer
	}

	return nil
}

// Answer is the representation of an API response answer.
type Answer struct {
	Status string `json:"status,omitempty"`
	Result bool   `json:"result,omitempty"`

	ErrorCode string `json:"error_code,omitempty"`
	ErrorText string `json:"error_text,omitempty"`
}

// ChangeRecordsRequest - data representation for data change request.
type ChangeRecordsRequest struct {
	Fqdn    string     `json:"fqdn,omitempty"`
	Records RecordList `json:"records,omitempty"`
}

// List of entries (in this case only described TXT)
type RecordList struct {
	TXT []TxtRecord `json:"TXT,omitempty"`
}

//Txt record structure
type TxtRecord struct {
	Priority int    `json:"priority,omitempty"`
	Value    string `json:"value,omitempty"`
}
