package internal

import "fmt"

type APIResponse struct {
	Succeeded bool     `json:"succeeded"`
	ErrorCode string   `json:"error_code"`
	ErrorMsg  string   `json:"error"`
	Domains   []Domain `json:"domains"`
}

func (a APIResponse) Error() string {
	return fmt.Sprintf("%s: %s", a.ErrorCode, a.ErrorMsg)
}

// Domain structure describes the config for an entire domain within Hover:
// the dates involved, contact addresses, nameservers, etc.:
// it seems to cover everything about the domain in one structure,
// which is convenient when you want to compare data across many domains.
type Domain struct {
	// A unique opaque identifier defined by Hover
	ID string `json:"id"`
	// the actual domain name.  ie: "example.com"
	DomainName string `json:"domain_name"`
	// DNS Records in a zone, if expanded.
	Records []Record `json:"entries,omitempty"`
}

// Record is a single DNS record, such as a single NS, TXT, A, PTR, AAAA record within a zone.
type Record struct {
	// A unique opaque identifier defined by Hover
	ID string `json:"id"`
	// seems to track the default @ or "*" record
	Default bool `json:"is_default"`
	// entry name, or "*" for default
	Name string `json:"name"`
	// TimeToLive, seconds
	TTL int `json:"ttl"`
	// record type: A, MX, PTR, TXT, etc
	Type string `json:"type"`
	// free-form text of verbatim value to store (ie "192.168.0.1" for A-rec)
	Content string `json:"content"`
}
