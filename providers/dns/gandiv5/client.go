package gandiv5

type deleteFieldRequest struct {
	Delete bool `json:"delete"`
}

type getFieldRequest struct {
	Get bool `json:"get"`
}

// types for JSON responses with only a message
type apiResponse struct {
	Message string `json:"message"`
}

// Record TXT record representation
type Record struct {
	RRSetTTL    int      `json:"rrset_ttl"`
	RRSetValues []string `json:"rrset_values"`
	RRSetName   string   `json:"rrset_name,omitempty"`
	RRSetType   string   `json:"rrset_type,omitempty"`
}
