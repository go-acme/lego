package internal

type RecordRequest struct {
	Zone    string
	Name    string
	Type    string
	Content string
}

type Record struct {
	ID      int    `json:"id,omitempty"`
	Content string `json:"content,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
}
