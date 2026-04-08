package internal

import "fmt"

type ChangeType string

const (
	ChangeTypeAdd     ChangeType = "ADD"
	ChangeTypeReplace ChangeType = "REPLACE"
	ChangeTypeDelete  ChangeType = "DELETE"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%d: %s", a.Code, a.Message)
}

type ZoneVersion struct {
	UUIDRef      string     `json:"uuid_ref"`
	Name         string     `json:"name"`
	CreationDate string     `json:"creation_date"`
	Active       bool       `json:"active"`
	Zone         *Reference `json:"zone"`
	Domain       *Reference `json:"domain"`
}

type Reference struct {
	Ref string `json:"$ref"`
}

type RecordRequest struct {
	Name     string `url:"name"`
	Type     string `url:"type"`
	Priority int    `url:"priority"`
	TTL      int    `url:"ttl"`
	Data     string `url:"data"`
}

type ResourceRecord struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Aux    int     `json:"aux"`
	TTL    int     `json:"ttl"`
	Data   string  `json:"data"`
	Domain *Domain `json:"domain"`
}

type Domain struct {
	ID       int        `json:"id"`
	Name     string     `json:"name"`
	DNSSec   bool       `json:"dnssec"`
	External bool       `json:"external"`
	Versions *Reference `json:"versions"`
	Zone     *Reference `json:"zone"`
}

type ResourceRecordOperation struct {
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	ChangeType ChangeType `json:"changeType"`
	Data       string     `json:"data,omitempty"`
	Records    []Record   `json:"records"`
}

type Record struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Priority int    `json:"priority,omitempty"`
	TTL      int    `json:"ttl"`
	Data     string `json:"data"`
}
