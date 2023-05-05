package internal

type apiResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// DNSRecord a DNS record.
type DNSRecord struct {
	ID            string        `json:"id,omitempty"`
	Type          string        `json:"type"`
	Value         any           `json:"value,omitempty"`
	Name          string        `json:"name,omitempty"`
	TTL           int           `json:"ttl,omitempty"`
	UpstreamHTTPS string        `json:"upstream_https,omitempty"`
	IPFilterMode  *IPFilterMode `json:"ip_filter_mode,omitempty"`
}

// TXTRecordValue represents a TXT record value.
type TXTRecordValue struct {
	Text string `json:"text,omitempty"` // only for TXT Record.
}

// IPFilterMode a DNS ip_filter_mode.
type IPFilterMode struct {
	Count     string `json:"count,omitempty"`
	Order     string `json:"order,omitempty"`
	GeoFilter string `json:"geo_filter,omitempty"`
}
