package internal

import "encoding/json"

// Record represents the content of a DNS record.
type Record struct {
	ID       int64  `json:"record_id,omitempty"`
	Name     string `json:"name,omitempty"`
	Data     string `json:"data,omitempty"`
	Type     string `json:"type,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

// apiResponse represents an API response.
type apiResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message"`
	Records json.RawMessage `json:"records,omitempty"`
	Record  json.RawMessage `json:"record,omitempty"`
}

type recordHeader struct {
	ID int64 `json:"id"`
}
