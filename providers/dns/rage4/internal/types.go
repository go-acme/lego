package internal

import (
	"fmt"
	"strings"
	"time"
)

type APIError struct {
	Status   int              `json:"status"`
	Type     string           `json:"type"`
	Title    string           `json:"title"`
	Instance string           `json:"instance"`
	TraceID  string           `json:"traceId"`
	Detail   string           `json:"detail"`
	Problems []ProblemDetails `json:"errors"`
}

func (e *APIError) Error() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "%d: %s", e.Status, e.Title)

	if e.Type != "" {
		_, _ = fmt.Fprintf(msg, ", type: %s", e.Type)
	}

	if e.Instance != "" {
		_, _ = fmt.Fprintf(msg, ", instance: %s", e.Instance)
	}

	if e.TraceID != "" {
		_, _ = fmt.Fprintf(msg, ", traceId: %s", e.TraceID)
	}

	if e.Detail != "" {
		_, _ = fmt.Fprintf(msg, ", detail: %s", e.Detail)
	}

	for _, pb := range e.Problems {
		_, _ = fmt.Fprintf(msg, ", %s", pb)
	}

	return msg.String()
}

type ProblemDetails struct {
	Name     string `json:"name"`
	Reason   string `json:"reason"`
	Code     string `json:"code"`
	Severity string `json:"severity"`
}

func (p ProblemDetails) String() string {
	return fmt.Sprintf("%s: %s: %s: %s", p.Code, p.Reason, p.Name, p.Severity)
}

type ServerErrorResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

func (e *ServerErrorResponse) Error() string {
	return fmt.Sprintf("%d: %s: %s: %s", e.Code, e.Status, e.Reason, e.Note)
}

type CommonResponse struct {
	Status bool   `json:"status"`
	ID     int    `json:"id"`
	Error  string `json:"error"`
}

type Record struct {
	DomainID         int    `json:"id,omitempty"`      // Required.
	Name             string `json:"name,omitempty"`    // Required.
	Content          string `json:"content,omitempty"` // Required.
	Type             string `json:"type,omitempty"`    // Required.
	TTL              int    `json:"ttl,omitempty"`
	Priority         int    `json:"priority,omitempty"`
	Weight           int    `json:"weight,omitempty"`
	Active           bool   `json:"isActive,omitempty"`
	FailoverEnabled  bool   `json:"isFailoverEnabled,omitempty"`
	FailoverContent  bool   `json:"failoverContent,omitempty"`
	FailoverWithdraw bool   `json:"isFailoverWithdraw,omitempty"`
	FailoverActive   bool   `json:"isFailoverActive,omitempty"`
	GeoZoneID        int    `json:"geoZoneID,omitempty"`
	GeoLatitude      int    `json:"geoLatitude,omitempty"`
	GeoLongitude     int    `json:"geoLongitude,omitempty"`
	GeoAsNum         int64  `json:"geoAsNum,omitempty"`
	UDPLimit         bool   `json:"udpLimit,omitempty"`
	Description      string `json:"description,omitempty"`
	WebhookID        int    `json:"webhookID,omitempty"`
}

type Domain struct {
	ID             int       `json:"id,omitempty"`
	Name           string    `json:"name,omitempty"`
	OwnerEmail     string    `json:"owner_email,omitempty"`
	DefaultNs1     string    `json:"default_ns1,omitempty"`
	DefaultNs2     string    `json:"default_ns2,omitempty"`
	Master         string    `json:"master,omitempty"`
	SoaRefresh     int       `json:"soa_refresh,omitempty"`
	SoaExpiry      int       `json:"soa_expiry,omitempty"`
	SoaRetry       int       `json:"soa_retry,omitempty"`
	SoaNx          int       `json:"soa_nx,omitempty"`
	NsTTL          int       `json:"ns_ttl,omitempty"`
	Online         bool      `json:"online,omitempty"`
	Secure         bool      `json:"secure,omitempty"`
	Locked         bool      `json:"locked,omitempty"`
	APIAccess      bool      `json:"api_access,omitempty"`
	Created        time.Time `json:"created,omitzero"`
	Updated        time.Time `json:"updated,omitzero"`
	SubnetMask     int       `json:"subnet_mask,omitempty"`
	Type           string    `json:"type,omitempty"`
	NotifiedSerial int       `json:"notified_serial,omitempty"`
	LastCheck      int       `json:"last_check,omitempty"`
}
