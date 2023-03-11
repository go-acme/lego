package internal

import "fmt"

type errorResponse struct {
	Status    string `json:"status"`
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
}

func (e errorResponse) Error() string {
	return fmt.Sprintf("status=%s, message=%s", e.Status, e.Message)
}

// DNSZone a DNS zone.
type DNSZone struct {
	ID             int    `json:"id"`
	Name           string `json:"name,omitempty"`
	Synced         bool   `json:"synced,omitempty"`
	QueriesCount   int    `json:"queries_count,omitempty"`
	RecordsCount   int    `json:"records_count,omitempty"`
	AliasesCount   int    `json:"aliases_count,omitempty"`
	RedirectsCount int    `json:"redirects_count,omitempty"`
	ForwardsCount  int    `json:"forwards_count,omitempty"`
	TemplateID     int    `json:"template_id,omitempty"`
}

// DNSRecord a DNS record.
type DNSRecord struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"`
	Content string `json:"content,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
	ZoneID  int    `json:"zone_id,omitempty"`
}
