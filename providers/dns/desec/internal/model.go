package internal

// RRSet DNS Record Set.
type RRSet struct {
	Name    string   `json:"name,omitempty"`
	Domain  string   `json:"domain,omitempty"`
	SubName string   `json:"subname,omitempty"`
	Type    string   `json:"type,omitempty"`
	Records []string `json:"records"`
	TTL     int      `json:"ttl,omitempty"`
}

// NotFound Not found error.
type NotFound struct {
	Detail string `json:"detail"`
}

func (n NotFound) Error() string {
	return n.Detail
}
