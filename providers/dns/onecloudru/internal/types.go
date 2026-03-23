package internal

type Domain struct {
	ID            int64    `json:"ID,omitempty"`
	Name          string   `json:"Name,omitempty"`
	TechName      string   `json:"TechName,omitempty"`
	State         string   `json:"State,omitempty"`
	IsDelegate    bool     `json:"IsDelegate,omitempty"`
	LinkedRecords []Record `json:"LinkedRecords,omitempty"`
}

type Record struct {
	ID                   int64  `json:"ID,omitempty"`
	Type                 string `json:"TypeRecord,omitempty"`
	IP                   string `json:"IP,omitempty"`
	HostName             string `json:"HostName,omitempty"`
	Priority             string `json:"Priority,omitempty"`
	Text                 string `json:"Text,omitempty"`
	Proto                string `json:"Proto,omitempty"`
	Service              string `json:"Service,omitempty"`
	Weight               string `json:"Weight,omitempty"`
	TTL                  int    `json:"TTL,omitempty"`
	CanonicalDescription string `json:"CanonicalDescription,omitempty"`
}

type CreateTXTRecordRequest struct {
	DomainID string `json:"DomainId"`
	Name     string `json:"Name"`
	TTL      string `json:"TTL"`
	Text     string `json:"Text"`
}
