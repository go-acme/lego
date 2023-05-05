package internal

type Recordset struct {
	RecordType string  `json:"type"`
	Name       string  `json:"name"`
	TTL        int     `json:"ttl,omitempty"`
	ID         string  `json:"id,omitempty"`
	Record     *Record `json:"record,omitempty"`
}

type Record struct {
	ID      string `json:"id,omitempty"`
	Content string `json:"content"`
	Enabled bool   `json:"enabled,omitempty"`
}

type Zone struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	DNSName string `json:"dnsName"`
	FQDN    string `json:"fqdn"`
	URI     string `json:"uri"`
}
