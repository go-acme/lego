package internal

// DNSRecord a DNS record.
type DNSRecord struct {
	ID            string       `json:"id,omitempty"`
	Name          string       `json:"name,omitempty"`
	Type          string       `json:"type,omitempty"`
	TTL           int          `json:"ttl,omitempty"`
	Value         interface{}  `json:"value"`
	UpstreamHTTPS string       `json:"upstream_https"`
	IPFilterMode  IPFilterMode `json:"ip_filter_mode"`
}

// DNSRecords a set of DNS record.
type DNSRecords struct {
	Records []DNSRecord `json:"data"`
}

// TxtValue a DNS value for txt record.
type TxtValue struct {
	Text string `json:"text"`
}

// IPFilterMode a DNS ip_filter_mode
type IPFilterMode struct {
	Count     string `json:"count"`
	Order     string `json:"order"`
	GeoFilter string `json:"geo_filter"`
}
