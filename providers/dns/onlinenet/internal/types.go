package internal

import "fmt"

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

type Record struct {
	ID       int    `json:"id" url:"-" `
	Name     string `json:"name" url:"name"`
	Type     string `json:"type" url:"type"`
	Priority int    `json:"priority" url:"priority"`
	TTL      int    `json:"ttl" url:"ttl"`
	Data     string `json:"data" url:"data"`
}
