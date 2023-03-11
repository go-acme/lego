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

type Response interface {
	GetStatus() int
	GetMessage() string
}

// apiResponse represents an API response.
type apiResponse[S any, R any] struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Records S      `json:"records,omitempty"`
	Record  R      `json:"record,omitempty"`
}

func (a apiResponse[S, R]) GetStatus() int {
	return a.Status
}

func (a apiResponse[S, R]) GetMessage() string {
	return a.Message
}

type recordHeader struct {
	ID int64 `json:"id"`
}
