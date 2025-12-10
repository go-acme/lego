package internal

// Zone represents a DNS zone.
type Zone struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	HumanName string `json:"human_name"`
}

// Record represents a DNS record.
type Record struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Content  string `json:"content,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"prio,omitempty"`
}

// RecordRequest is the request body for creating/updating a record.
type RecordRequest struct {
	Record Record `json:"record"`
}
