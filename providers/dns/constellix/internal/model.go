package internal

import (
	"strings"
)

// APIError is the representation of an API error.
type APIError struct {
	Errors []string `json:"errors"`
}

func (a APIError) Error() string {
	return strings.Join(a.Errors, ": ")
}

// SuccessMessage is the representation of a success message.
type SuccessMessage struct {
	Success string `json:"success"`
}

// RecordRequest is the representation of a request's record.
type RecordRequest struct {
	Name       string        `json:"name"`
	TTL        int           `json:"ttl,omitempty"`
	RoundRobin []RecordValue `json:"roundRobin,omitempty"`
}

// RecordValue is the representation of a record's value.
type RecordValue struct {
	Value       string `json:"value,omitempty"`
	DisableFlag bool   `json:"disableFlag,omitempty"` // only for the response
}

// Record is the representation of a record.
type Record struct {
	ID           int64         `json:"id"`
	Type         string        `json:"type"`
	RecordType   string        `json:"recordType"`
	Name         string        `json:"name"`
	RecordOption string        `json:"recordOption,omitempty"`
	NoAnswer     bool          `json:"noAnswer,omitempty"`
	Note         string        `json:"note,omitempty"`
	TTL          int           `json:"ttl,omitempty"`
	GtdRegion    int           `json:"gtdRegion,omitempty"`
	ParentID     int           `json:"parentId,omitempty"`
	Parent       string        `json:"parent,omitempty"`
	Source       string        `json:"source,omitempty"`
	ModifiedTs   int64         `json:"modifiedTs,omitempty"`
	Value        []RecordValue `json:"value,omitempty"`
	RoundRobin   []RecordValue `json:"roundRobin,omitempty"`
}

// Domain is the representation of a domain.
type Domain struct {
	ID      int64  `json:"id"`
	Name    string `json:"name,omitempty"`
	TypeID  int64  `json:"typeId,omitempty"`
	Version int64  `json:"version,omitempty"`
	Status  string `json:"status,omitempty"`
}

// PaginationParameters is pagination parameters.
type PaginationParameters struct {
	// Offset retrieves a subset of records starting with the offset value.
	Offset int `url:"offset"`
	// Max retrieves maximum number of dataset.
	Max int `url:"max"`
	// Sort on the basis of given property name.
	Sort string `url:"sort"`
	// Order Sort order. Possible values are asc / desc.
	Order string `url:"order"`
}
