package gandiv5

// types for JSON method calls and parameters
type addFieldRequest struct {
	RRSetTTL    int      `json:"rrset_ttl"`
	RRSetValues []string `json:"rrset_values"`
}

type deleteFieldRequest struct {
	Delete bool `json:"delete"`
}

type getFieldRequest struct {
	Get bool `json:"get"`
}

// types for JSON responses with only a message
type responseMessage struct {
	Message string `json:"message"`
}

// types for JSON responses with a struct
type responseStruct struct {
	RRSetTTL    int      `json:"rrset_ttl"`
	RRSetValues []string `json:"rrset_values"`
	RRSetName   string   `json:"rrset_name"`
	RRSetType   string   `json:"rrset_type"`
}
