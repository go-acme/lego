package internal

// RecordBody represents the content of a DNS record.
type RecordBody struct {
	Name     string `json:"name"`
	Data     string `json:"data"`
	Type     string `json:"type"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority"`
}

// Record represents a concrete DNS record.
type Record struct {
	ID       int64  `json:"record_id"`
	Name     string `json:"name"`
	Data     string `json:"data"`
	Type     string `json:"type"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority"`
}

// RecordHeader represents the identifying info of a concrete DNS record.
type RecordHeader struct {
	ID int64 `json:"id"`
}
