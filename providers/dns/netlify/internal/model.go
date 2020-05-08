package internal

// DNSRecord DNS record representation.
type DNSRecord struct {
	ID       string `json:"id,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value,omitempty"`
}
