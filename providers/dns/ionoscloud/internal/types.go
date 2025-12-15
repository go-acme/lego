package internal

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type APIError struct {
	HTTPStatus int            `json:"httpStatus"`
	Messages   []ErrorMessage `json:"messages"`
}

func (a *APIError) Error() string {
	var msg strings.Builder

	msg.WriteString(strconv.Itoa(a.HTTPStatus))

	for _, m := range a.Messages {
		msg.WriteString(": ")
		msg.WriteString(m.String())
	}

	return msg.String()
}

type ErrorMessage struct {
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
}

func (e ErrorMessage) String() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.Message)
}

type ZonesResponse struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Items  []Zone `json:"items"`
}

type Zone struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Metadata   ZoneMetadata   `json:"metadata"`
	Properties ZoneProperties `json:"properties"`
}

type ZoneMetadata struct {
	CreatedDate          time.Time `json:"createdDate"`
	CreatedBy            string    `json:"createdBy"`
	CreatedByUserID      string    `json:"createdByUserId"`
	LastModifiedDate     time.Time `json:"lastModifiedDate"`
	LastModifiedBy       string    `json:"lastModifiedBy"`
	LastModifiedByUserID string    `json:"lastModifiedByUserId"`
	ResourceURN          string    `json:"resourceURN"`
	State                string    `json:"state"`
	Nameservers          []string  `json:"nameservers"`
}

type ZoneProperties struct {
	ZoneName    string `json:"zoneName"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type RecordResponse struct {
	ID         string           `json:"id"`
	Type       string           `json:"type"`
	Metadata   RecordMetadata   `json:"metadata"`
	Properties RecordProperties `json:"properties"`
}

type RecordMetadata struct {
	CreatedDate          time.Time `json:"createdDate"`
	CreatedBy            string    `json:"createdBy"`
	CreatedByUserID      string    `json:"createdByUserId"`
	LastModifiedDate     time.Time `json:"lastModifiedDate"`
	LastModifiedBy       string    `json:"lastModifiedBy"`
	LastModifiedByUserID string    `json:"lastModifiedByUserId"`
	ResourceURN          string    `json:"resourceURN"`
	State                string    `json:"state"`
	Fqdn                 string    `json:"fqdn"`
	ZoneID               string    `json:"zoneId"`
}

type RecordProperties struct {
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
	Content  string `json:"content,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
}
