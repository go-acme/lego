package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// APIError is the error response of the API.
type APIError struct {
	Code    int             `json:"error"`
	Message string          `json:"message"`
	Meta    json.RawMessage `json:"meta,omitempty"`
}

func (a *APIError) Error() string {
	msg := new(bytes.Buffer)

	_, _ = fmt.Fprintf(msg, "%d: %s", a.Code, a.Message)

	if len(a.Meta) > 0 && string(a.Meta) != "null" {
		msg.WriteString(": ")

		err := json.Compact(msg, a.Meta)
		if err != nil {
			msg.WriteString(strings.NewReplacer("\n", "", "\r", "").Replace(string(a.Meta)))
		}
	}

	return msg.String()
}

// APIResponse is the common envelope of the API responses.
type APIResponse[T any] struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// ListResponse is the common envelope of the paginated API responses.
type ListResponse[T any] struct {
	APIResponse[[]T]

	Meta Pagination `json:"meta"`
}

type Pagination struct {
	Total   int `json:"total"`
	Count   int `json:"count"`
	Pages   int `json:"pages"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type Purchase struct {
	PurchaseID   int          `json:"purchase_id"`
	DomainID     int          `json:"domain_id"`
	PurchaseName PurchaseName `json:"purchase_name"`
}

type PurchaseName struct {
	Name       string `json:"name"`
	FullName   string `json:"full_name"`
	ObjectName string `json:"object_name"`
}

type Domain struct {
	DomainID   int        `json:"domain_id"`
	DomainName DomainName `json:"domain_name"`
}

type DomainName struct {
	Unicode  string `json:"unicode"`
	Punycode string `json:"punycode"`
}

type Record struct {
	ID          string `json:"record_id,omitempty"`
	Type        string `json:"type,omitempty"`
	Host        string `json:"host,omitempty"`
	Destination string `json:"destination,omitempty"`
}
