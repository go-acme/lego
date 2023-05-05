package internal

// Record represents a record.
type Record struct {
	Hostname string `url:"hostname,omitempty"`
	Type     string `url:"type,omitempty"`
	Value    string `url:"value,omitempty"`
	TTL      int    `url:"ttl,omitempty"`
}
