package internal

type APIRRSet struct {
	DNSZoneName string `json:"dns_zone_name,omitempty"`
	GroupName   string `json:"group_name,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	RecordName  string `json:"record_name,omitempty"`
	Type        string `json:"type,omitempty"`
	RRSet       RRSet  `json:"rrset,omitempty"`
}

type RRSetRequest struct {
	DNSZoneName string `json:"dns_zone_name,omitempty"`
	GroupName   string `json:"group_name,omitempty"`
	RRSet       RRSet  `json:"rrset,omitempty"`
}

type RRSet struct {
	Description string     `json:"description,omitempty"`
	TTL         int        `json:"ttl,omitempty"`
	TXTRecord   *TXTRecord `json:"txt_record,omitempty"`
}

type TXTRecord struct {
	Name   string   `json:"name,omitempty"`
	Values []string `json:"values,omitempty"`
}
