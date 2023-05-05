package internal

// Some fields have been omitted from the structs
// because they are not required for this application.

type DomainListingResponse struct {
	Page     int                `json:"page"`
	Limit    int                `json:"limit"`
	Pages    int                `json:"pages"`
	Total    int                `json:"total"`
	Embedded EmbeddedDomainList `json:"_embedded"`
}

type EmbeddedDomainList struct {
	Domains []*Domain `json:"domains"`
}

type Domain struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type DomainResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Created string `json:"created"`
	PaidUp  string `json:"payed_up"`
	Active  bool   `json:"active"`
}

type NameserverResponse struct {
	General     NameserverGeneral `json:"general"`
	Nameservers []*Nameserver     `json:"nameservers"`
	SOA         NameserverSOA     `json:"soa"`
}

type NameserverGeneral struct {
	IPv4       string `json:"ip_v4"`
	IPv6       string `json:"ip_v6"`
	IncludeWWW bool   `json:"include_www"`
}

type NameserverSOA struct {
	Mail    string `json:"mail"`
	Refresh int    `json:"refresh"`
	Retry   int    `json:"retry"`
	Expiry  int    `json:"expiry"`
	TTL     int    `json:"ttl"`
}

type Nameserver struct {
	Name string `json:"name"`
}

type RecordListingResponse struct {
	Page     int                `json:"page"`
	Limit    int                `json:"limit"`
	Pages    int                `json:"pages"`
	Total    int                `json:"total"`
	Embedded EmbeddedRecordList `json:"_embedded"`
}

type EmbeddedRecordList struct {
	Records []*Record `json:"records"`
}

type Record struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority"`
	Type     string `json:"type"`
}
