package internal

import (
	"fmt"
	"strings"
)

type RecordsResponse struct {
	Data []Record `json:"records,omitempty"`
}

type NameserverRecordPayload struct {
	Data Record `json:"nameserver_record"`
}

type DomainsResponse struct {
	Data []Domain `json:"domains,omitempty"`
}

type APIResponse struct {
	Status  string              `json:"status,omitempty"`
	Details map[string][]string `json:"errors,omitempty"`
}

func (a APIResponse) Error() string {
	var details []string
	for k, v := range a.Details {
		details = append(details, fmt.Sprintf("%s: %s", k, v))
	}

	return fmt.Sprintf("status: %s, details: %s", a.Status, strings.Join(details, ","))
}

type Record struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
	Type    string `json:"type,omitempty"`
}

type Domain struct {
	ID          int    `json:"id,omitempty"`
	UnicodeFqdn string `json:"unicode_fqdn,omitempty"`
	Domain      string `json:"domain,omitempty"`
	TLD         string `json:"tld,omitempty"`
	Status      string `json:"status,omitempty"`
}
