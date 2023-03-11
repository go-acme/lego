package internal

import "fmt"

// Zones is the response struct from the Stackpath api GetZones.
type Zones struct {
	Zones []Zone `json:"zones"`
}

// Zone a DNS zone representation.
type Zone struct {
	ID     string
	Domain string
}

// Records is the response struct from the Stackpath api GetZoneRecords.
type Records struct {
	Records []Record `json:"records"`
}

// Record a DNS record representation.
type Record struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  int    `json:"ttl"`
	Data string `json:"data"`
}

// ErrorResponse the API error response representation.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Message)
}
