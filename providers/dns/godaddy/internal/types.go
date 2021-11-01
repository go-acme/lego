package internal

// DNSRecord a DNS record.
type DNSRecord struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	Data string `json:"data"`
	TTL  int    `json:"ttl,omitempty"`

	Priority int    `json:"priority,omitempty"`
	Port     int    `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Service  string `json:"service,omitempty"`
	Weight   int    `json:"weight,omitempty"`
}
