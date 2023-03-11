package internal

import "encoding/json"

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Data Domain information.
type Data struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
	TTL    int    `json:"ttl,omitempty"`
}

// TXTRecord a TXT record.
type TXTRecord struct {
	ID       int    `json:"domain_id,omitempty"`
	RecordID string `json:"record_id,omitempty"`

	Host   string `json:"host"`
	Value  string `json:"value"`
	Type   string `json:"type"`
	LineID int    `json:"line_id,string"`
	TTL    int    `json:"ttl,string"`
}
