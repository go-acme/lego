package selectel

import "fmt"

// Domain represents domain name.
type Domain struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Record represents DNS record.
type Record struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"` // Record type (SOA, NS, A/AAAA, CNAME, SRV, MX, TXT, SPF)
	TTL     int    `json:"ttl,omitempty"`
	Email   string `json:"email,omitempty"`   // Email of domain's admin (only for SOA records)
	Content string `json:"content,omitempty"` // Record content (not for SRV)
}

// APIError API error message
type APIError struct {
	Description string `json:"error"`
	Code        int    `json:"code"`
	Field       string `json:"field"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("API error: %d - %s - %s", a.Code, a.Description, a.Field)
}
