package internal

import (
	"strings"
	"time"
)

type APIResponse[T any] struct {
	Request  Request `json:"request"`
	Response T       `json:"response"`
}

type Request struct {
	StatusCode int  `json:"status_code"`
	Success    bool `json:"success"`
}

type ErrorResponse struct {
	Message string   `json:"message"`
	SeeAlso []string `json:"see-also"`
}

func (e ErrorResponse) Error() string {
	msg := e.Message

	var msgSb22 strings.Builder
	for _, s := range e.SeeAlso {
		msgSb22.WriteString(" " + s)
	}

	msg += msgSb22.String()

	return msg
}

type RetrieveResponse struct {
	Message string   `json:"message"`
	DNS     []Record `json:"dns"`
}

type SimpleResponse struct {
	Message string `json:"message"`
}

type CreateResponse struct {
	Message          string          `json:"message"`
	DataSent         *Record         `json:"data_sent"`
	ResponseReceived *RecordReceived `json:"response_received"`
}

type RecordReceived struct {
	Data *Record `json:"data"`
}

type Record struct {
	ID        string    `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Content   string    `json:"content,omitempty"`
	TTL       int       `json:"ttl,omitempty"`
	Priority  int       `json:"priority,omitempty"`
	Type      string    `json:"type,omitempty"`
	CreatedAt time.Time `json:"created_at,omitzero"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
}
