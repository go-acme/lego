package internal

import "fmt"

type BaseResponse struct {
	Code        int    `json:"code"`
	Description string `json:"desc"`
}

func (r BaseResponse) Check() error {
	// Response codes:
	// - 1000: Command completed successfully
	// - 1300: Command completed successfully; no messages
	// - 2001: Command syntax error
	// - 2002: Command use error
	// - 2003: Required parameter missing
	// - 2004: Parameter value range error
	// - 2104: Billing failure
	// - 2200: Authentication error
	// - 2201: Authorization error
	// - 2303: Object does not exist
	// - 2304: Object status prohibits operation
	// - 2309: Object duplicate found
	// - 2400: Command failed
	// - 2500: Command failed; server closing connection
	if r.Code != 1000 && r.Code != 1300 {
		return fmt.Errorf("%d: %s", r.Code, r.Description)
	}

	return nil
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
