package internal

type Record struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Prefix   string `json:"prefix,omitempty"`
	Content  string `json:"content,omitempty"`
	Priority int    `json:"prio,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Active   bool   `json:"active,omitempty"`
}
