package internal

import "fmt"

type APIError struct {
	Code int    `json:"code,omitempty"`
	Desc string `json:"desc,omitempty"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s (code: %d)", a.Desc, a.Code)
}

type APIResponse[T any] struct {
	Code int    `json:"code,omitempty"`
	Desc string `json:"desc,omitempty"`
	Data T
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IP       string `json:"ip"`
}

type Token struct {
	Token      string `json:"token"`
	ResellerID int    `json:"reseller_id"`
}

type Results[T any] struct {
	Results []T `json:"results"`
	Total   int `json:"total"`
}

type Zone struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Records    []Record `json:"records"`
	ResellerID int      `json:"reseller_id"`
	Type       string   `json:"type"`
}

type Record struct {
	Name  string `json:"name"`
	TTL   int    `json:"ttl"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ZoneAction struct {
	ID      int          `json:"id"`
	Name    string       `json:"name"`
	Records RecordAction `json:"records"`
}

type RecordAction struct {
	Add     []Record       `json:"add,omitempty"`
	Update  []UpdateAction `json:"update,omitempty"`
	Remove  []Record       `json:"remove,omitempty"`
	Replace []Record       `json:"replace,omitempty"`
}

type UpdateAction struct {
	OriginalRecord Record `json:"original_record"`
	Record         Record `json:"record"`
}

type ResponseSuccess struct {
	Success bool `json:"success"`
}

type ZonesRequest struct {
	Limit  int `url:"limit,omitempty"`
	Offset int `url:"offset,omitempty"`

	ByCreationDate     string `url:"order_by.creation_date,omitempty"`
	ByModificationDate string `url:"order_by.modification_date,omitempty"`
	ByName             string `url:"order_by.name,omitempty"`

	NamePattern string `url:"name_pattern,omitempty"`
}
