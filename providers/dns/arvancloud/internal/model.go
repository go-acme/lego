package internal

// DNSRecordTextValue represents a TXT record value.
type DNSRecordTextValue struct {
	Text string `json:"text,omitempty"`
}

// DNSRecord a DNS record.
type DNSRecord struct {
	ID    string      `json:"id,omitempty"`
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
	Name  string      `json:"name,omitempty"`
	TTL   int         `json:"ttl,omitempty"`
}

// TxtDNSRecord a DNS record.
type TxtDNSRecord struct {
	DNSRecord
	Value DNSRecordTextValue `json:"value,omitempty"`
}

// DNSRecords a set of DNS record.
type DNSRecords struct {
	Records []DNSRecord `json:"data,omitempty"`
}
