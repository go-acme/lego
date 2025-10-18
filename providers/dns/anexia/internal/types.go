package internal

import "fmt"

type APIError struct {
	Details struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%d: %s", a.Details.Code, a.Details.Message)
}

type Zone struct {
	Identifier string     `json:"identifier,omitempty"`
	Name       string     `json:"name,omitempty"`
	TTL        int        `json:"ttl,omitempty"`
	ZoneName   string     `json:"zone_name,omitempty"`
	Revisions  []Revision `json:"revisions,omitempty"`
}

type Revision struct {
	Identifier string   `json:"identifier,omitempty"`
	Records    []Record `json:"records,omitempty"`
	State      string   `json:"state,omitempty"`
}

type Record struct {
	Identifier string `json:"identifier,omitempty"`
	Immutable  bool   `json:"immutable,omitempty"`
	Name       string `json:"name,omitempty"`
	RData      string `json:"rdata,omitempty"`
	Region     string `json:"region"`
	TTL        int    `json:"ttl,omitempty"`
	Type       string `json:"type,omitempty"`
}
