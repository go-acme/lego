package shared

type Record struct {
	DName      string   `json:"dname,omitempty"`
	TTL        int      `json:"ttl,omitempty"`
	RecordType string   `json:"record_type,omitempty"`
	Data       []string `json:"data,omitempty"`
	LineIndex  int      `json:"line_index,omitempty"`
}

type ZoneRecord struct {
	LineIndex  int      `json:"line_index,omitempty"`
	Type       string   `json:"type,omitempty"`
	DataB64    []string `json:"data_b64,omitempty"`
	DNameB64   string   `json:"dname_b64,omitempty"`
	TextB64    string   `json:"text_b64,omitempty"`
	RecordType string   `json:"record_type,omitempty"`
	TTL        int      `json:"ttl,omitempty"`
}

type ZoneSerial struct {
	NewSerial string `json:"new_serial,omitempty"`
}
