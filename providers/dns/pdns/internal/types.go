package internal

type Record struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`

	// pre-v1 API
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  int    `json:"ttl,omitempty"`
}

type HostedZone struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	URL    string  `json:"url"`
	Kind   string  `json:"kind"`
	RRSets []RRSet `json:"rrsets"`

	// pre-v1 API
	Records []Record `json:"records"`
}

type RRSet struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Kind       string   `json:"kind"`
	ChangeType string   `json:"changetype"`
	Records    []Record `json:"records,omitempty"`
	TTL        int      `json:"ttl,omitempty"`
}

type RRSets struct {
	RRSets []RRSet `json:"rrsets"`
}

type apiError struct {
	ShortMsg string `json:"error"`
}

func (a apiError) Error() string {
	return a.ShortMsg
}

type apiVersion struct {
	URL     string `json:"url"`
	Version int    `json:"version"`
}
