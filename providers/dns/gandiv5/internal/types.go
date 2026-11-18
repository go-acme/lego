package internal

// types for JSON responses with only a message.
type apiResponse struct {
	Message string `json:"message"`
	UUID    string `json:"uuid,omitempty"`
}

// Record TXT record representation.
type Record struct {
	TTL    int      `json:"rrset_ttl,omitempty"`
	Values []string `json:"rrset_values"`
	Name   string   `json:"rrset_name,omitempty"`
	Type   string   `json:"rrset_type,omitempty"`
}
