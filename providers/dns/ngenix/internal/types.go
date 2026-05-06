package internal

import "fmt"

// APIError represents an API error response.
type APIError struct {
	Code   int    `json:"code"`
	Detail string `json:"detail"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", a.Code, a.Detail)
}

// DNSZoneCollection is the response for listing DNS zones.
type DNSZoneCollection struct {
	Elements []DNSZoneCollectionView `json:"elements"`
}

// DNSZoneCollectionView is a brief representation of a DNS zone.
type DNSZoneCollectionView struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// DNSZone is the full representation of a DNS zone.
type DNSZone struct {
	ID      int64       `json:"id,omitempty"`
	Name    string      `json:"name,omitempty"`
	Records []DNSRecord `json:"records"`
	Comment string      `json:"comment,omitempty"`
	DNSSec  *DNSSec     `json:"dnssec,omitempty"`
}

type DNSSec struct {
	Enabled      bool        `json:"enabled"`
	DNSSecKeyRef *Identifier `json:"dnssecKeyRef,omitempty"`
}

type Identifier struct {
	ID int64 `json:"id"`
}

// DNSZoneUpdate is the request body for updating a DNS zone.
type DNSZoneUpdate struct {
	Records []DNSRecord `json:"records"`
	Comment string      `json:"comment,omitempty"`
	DNSSec  *DNSSec     `json:"dnssec,omitempty"`
}

// DNSRecord represents a DNS record.
type DNSRecord struct {
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	Data           string      `json:"data,omitempty"`
	ConfigRef      *Identifier `json:"configRef,omitempty"`
	TargetGroupRef *Identifier `json:"targetGroupRef,omitempty"`
}
