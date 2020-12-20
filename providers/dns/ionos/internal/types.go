package internal

import "fmt"

// Zone defines model for zone.
type Zone struct {

	// The zone id.
	ID string `json:"id,omitempty"`

	// The zone name.
	Name string `json:"name,omitempty"`

	// Represents the possible zone types.
	Type string `json:"type,omitempty"`
}

// CustomerZone defines model for customer-zone.
type CustomerZone struct {

	// The zone id.
	ID string `json:"id,omitempty"`

	// The zone name
	Name    string   `json:"name,omitempty"`
	Records []Record `json:"records,omitempty"`

	// Represents the possible zone types.
	Type string `json:"type,omitempty"`
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

// Errors defines model for errors.
type Errors []struct {
	ErrorRecord    *Record  `json:"errorRecord,omitempty"`
	InputRecord    *Record  `json:"inputRecord,omitempty"`
	Invalid        []string `json:"invalid,omitempty"`
	InvalidFields  []string `json:"invalidFields,omitempty"`
	RequiredFields []string `json:"requiredFields,omitempty"`
}

// Record defines model for record.
type Record struct {
	ID string `json:"id,omitempty"`

	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`

	// Time to live for the record, recommended 3600.
	TTL int `json:"ttl,omitempty"`

	// Holds supported dns record types.
	Type string `json:"type,omitempty"`

	Priority int `json:"prio,omitempty"`

	// When is true, the record is not visible for lookup.
	Disabled bool `json:"disabled,omitempty"`
}

type RecordsFilter struct {
	// The FQDN used to filter all the record names that end with it.
	Suffix string `url:"suffix,omitempty"`

	// The record names that should be included (same as name field of Record)
	RecordName string `url:"recordName,omitempty"`

	// A comma-separated list of record types that should be included
	RecordType string `url:"recordType,omitempty"`
}
