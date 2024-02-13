package internal

import "fmt"

type ZoneRequest struct {
	Data Zone `json:"data"`
}

type Pagination struct {
	Offset int `json:"offset,omitempty" url:"offset"`
	Limit  int `json:"limit,omitempty" url:"limit"`
	Total  int `json:"total,omitempty" url:"total"`
}

type APIResponse[T any] struct {
	Meta  *Metadata `json:"meta,omitempty"`
	Data  T         `json:"data,omitempty"`
	Error *Error    `json:"error,omitempty"`
}

type Metadata struct {
	*Pagination

	InvocationID string `json:"invocationId,omitempty"`
}

type Zone struct {
	Type       string     `json:"type,omitempty"`
	ID         string     `json:"id,omitempty"`
	Attributes Attributes `json:"attributes"`
}

type Attributes struct {
	OrganisationID          string                         `json:"organisationId,omitempty"`
	OrganisationDescription string                         `json:"organisationDescription,omitempty"`
	DNSTypeDescription      string                         `json:"dnsTypeDescription,omitempty"`
	Slave                   bool                           `json:"slave,omitempty"`
	Pending                 bool                           `json:"pending,omitempty"`
	Deleted                 bool                           `json:"deleted,omitempty"`
	Settings                *Settings                      `json:"settings,omitempty"`
	Records                 map[string]map[string][]Record `json:"records,omitempty"`
	Redirects               map[string][]Redirect          `json:"redirects,omitempty"`
}

type Settings struct {
	MName   string `json:"mname,omitempty"`
	Refresh int    `json:"refresh,omitempty"`
	Expire  int    `json:"expire,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
}

type Record struct {
	TTL      int    `json:"ttl,omitempty"`
	Data     string `json:"rdata,omitempty"`
	Comments string `json:"comments,omitempty"`
}

type Redirect struct {
	Path        string `json:"path"`
	Destination string `json:"destination"`
	Status      int    `json:"status"`
	Slugs       bool   `json:"slugs"`
	Certificate bool   `json:"certificate"`
}

type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("api error: status=%d, message=%s", e.Status, e.Message)
}
