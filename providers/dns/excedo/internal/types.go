package internal

type BaseResponse struct {
	Code        int    `json:"code"`
	Description string `json:"desc"`
}

type GetRecordsResponse struct {
	BaseResponse

	DNS map[string]Zone `json:"dns"`
}

type Zone struct {
	DNSType string   `json:"dnstype"`
	Records []Record `json:"records"`
}

type Record struct {
	DomainName string `json:"domainName,omitempty" url:"domainName,omitempty"`
	RecordID   string `json:"recordid,omitempty" url:"recordid,omitempty"`
	Name       string `json:"name,omitempty" url:"name,omitempty"`
	Type       string `json:"type,omitempty" url:"type,omitempty"`
	Content    string `json:"content,omitempty" url:"content,omitempty"`
	TTL        string `json:"ttl,omitempty" url:"ttl,omitempty"`
}

type AddRecordResponse struct {
	BaseResponse

	RecordID int64 `json:"recordid"`
}

type LoginResponse struct {
	BaseResponse

	Parameters struct {
		Token string `json:"token"`
	} `json:"parameters"`
}
