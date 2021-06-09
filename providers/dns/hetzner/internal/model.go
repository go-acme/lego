package internal

// DNSRecord a DNS record.
type DNSRecord struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value"`
	Priority int    `json:"priority,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	ZoneID   string `json:"zone_id,omitempty"`
}

// DNSRecords a set of DNS record.
type DNSRecords struct {
	Records []DNSRecord `json:"records"`
}

// Zone a DNS zone.
type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Zones a set of DNS zones.
type Zones struct {
	Zones []Zone `json:"zones"`
}
