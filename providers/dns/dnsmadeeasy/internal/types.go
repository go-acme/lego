package internal

// Domain holds the DNSMadeEasy API representation of a Domain.
type Domain struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Record holds the DNSMadeEasy API representation of a Domain Record.
type Record struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	SourceID int    `json:"sourceId"`
}

type recordsResponse struct {
	Records *[]Record `json:"data"`
}
