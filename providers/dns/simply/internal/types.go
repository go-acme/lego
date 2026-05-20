package internal

// Record represents the content of a DNS record.
type Record struct {
	ID       int64  `json:"record_id,omitempty"`
	Name     string `json:"name,omitempty"`
	Data     string `json:"data,omitempty"`
	Type     string `json:"type,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

// apiResponse represents an API response payload. The wire format also
// includes "status" and "message" fields, but they are deprecated per the
// OpenAPI spec ("will be removed in a future API version") and intentionally
// not parsed.
type apiResponse[S any, R any] struct {
	Records S `json:"records,omitempty"`
	Record  R `json:"record,omitempty"`
}

type recordHeader struct {
	ID int64 `json:"id"`
}
