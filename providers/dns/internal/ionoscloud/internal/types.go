package internal

import (
	"fmt"
	"strconv"
	"strings"
)

// ClientError a detailed error.
type ClientError struct {
	errors     []Error
	StatusCode int
	message    string
}

func (f ClientError) Error() string {
	var msg strings.Builder

	msg.WriteString(strconv.Itoa(f.StatusCode) + ": ")

	if f.message != "" {
		msg.WriteString(f.message + ": ")
	}

	for i, e := range f.errors {
		if i != 0 {
			msg.WriteString(", ")
		}

		msg.WriteString(e.Error())
	}

	return msg.String()
}

func (f ClientError) Unwrap() error {
	if len(f.errors) == 0 {
		return nil
	}

	return &f.errors[0]
}

// Error defines model for error.
type Error struct {
	// The error code.
	Code string `json:"code,omitempty"`

	// The error message.
	Message string `json:"message,omitempty"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Zone defines model for zone.
type Zone struct {
	// The zone id.
	ID string `json:"id,omitempty"`

	// Represents the possible zone types.
	Type string `json:"type,omitempty"`

	Properties struct {
		ZoneName    string `json:"zoneName"`
		Description string `json:"description,omitempty"`
		Enabled     bool   `json:"enabled"`
	} `json:"properties"`
}

// Record defines model for record.
type Record struct {
	ID       string `json:"id,omitempty"`
	MetaData struct {
		FQDN string `json:"fqdn,omitempty"`
	} `json:"metadata,omitzero"`
	Properties RecordProperties `json:"properties"`
}

type RecordProperties struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`

	// Time to live for the record, recommended 3600.
	TTL int `json:"ttl,omitempty"`

	// Holds supported dns record types.
	Type string `json:"type,omitempty"`

	Priority int `json:"priority,omitempty"`

	// When is true, the record is not visible for lookup.
	Enabled bool `json:"enabled,omitempty"`
}
