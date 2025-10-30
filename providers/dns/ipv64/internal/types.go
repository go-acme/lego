package internal

import "fmt"

type APIResponse struct {
	Status string `json:"status"`
	Info   string `json:"info"`
}

// error

type APIError struct {
	APIResponse

	AddRecordMessage string `json:"add_record"`
	DelRecordMessage string `json:"del_record"`
	AddDomainMessage string `json:"add_domain"`
	DelDomainMessage string `json:"del_domain"`
}

func (a APIError) Error() string {
	msg := a.Info
	switch {
	case a.AddRecordMessage != "":
		msg = a.AddRecordMessage
	case a.DelRecordMessage != "":
		msg = a.DelRecordMessage
	case a.AddDomainMessage != "":
		msg = a.AddDomainMessage
	case a.DelDomainMessage != "":
		msg = a.DelDomainMessage
	}

	if msg == "" {
		return fmt.Sprintf("%s: %s", a.Status, a.Info)
	}

	return fmt.Sprintf("%s (%s): %s", a.Info, a.Status, msg)
}

// get_domains

type Domains struct {
	APIResponse

	APICall    string               `json:"add_domain"`
	Subdomains map[string]Subdomain `json:"subdomains"`
}

type Subdomain struct {
	Updates          int      `json:"updates"`
	Wildcard         int      `json:"wildcard"`
	DomainUpdateHash string   `json:"domain_update_hash"`
	Records          []Record `json:"records"`
}

type Record struct {
	RecordID   int    `json:"record_id"`
	Content    string `json:"content"`
	TTL        int    `json:"ttl"`
	Type       string `json:"type"`
	Prefix     string `json:"praefix"`
	LastUpdate string `json:"last_update"`
	RecordKey  string `json:"record_key"`
}
