package internal

type APIError struct {
	Message string `json:"error"`
}

func (e *APIError) Error() string {
	return e.Message
}

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
type apiResponse[S any, R any] struct {
	Records S `json:"records,omitempty"`
	Record  R `json:"record,omitempty"`
}

type recordHeader struct {
	ID int64 `json:"id"`
}
