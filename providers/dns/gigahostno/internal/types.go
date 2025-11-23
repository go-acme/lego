package internal

import "fmt"

type APIError struct {
	Meta MetaData `json:"meta"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%d: %s: %s", a.Meta.Status, a.Meta.StatusMessage, a.Meta.Message)
}

type MetaData struct {
	Status        int    `json:"status,omitempty"`
	StatusMessage string `json:"status_message,omitempty"`
	Maintenance   bool   `json:"maintenance"`
	Message       string `json:"message,omitempty"`
}

type APIResponse[T any] struct {
	Meta MetaData `json:"meta"`
	Data T        `json:"data,omitempty"`
}

type Zone struct {
	ZoneID           string `json:"zone_id,omitempty"`
	CustomerID       string `json:"cust_id,omitempty"`
	ZoneName         string `json:"zone_name,omitempty"`
	ZoneNameDisplay  string `json:"zone_name_display,omitempty"`
	ZoneType         string `json:"zone_type,omitempty"`
	ZoneActive       string `json:"zone_active,omitempty"`
	ZoneProtected    string `json:"zone_protected,omitempty"`
	ZoneIsRegistered string `json:"zone_is_registered,omitempty"`
	DomainRegistrar  string `json:"domain_registrar,omitempty"`
	DomainStatus     string `json:"domain_status,omitempty"`
	DomainExpiryDate string `json:"domain_expiry_date,omitempty"`
	DomainAutoRenew  string `json:"domain_auto_renew,omitempty"`
	ExternalDNS      string `json:"external_dns,omitempty"`
	RecordCount      int    `json:"record_count,omitempty"`
	ZoneUpdated      int    `json:"zone_updated,omitempty"`
}

type Record struct {
	RecordID    string `json:"record_id,omitempty"`
	RecordName  string `json:"record_name,omitempty"`
	RecordType  string `json:"record_type,omitempty"`
	RecordValue string `json:"record_value,omitempty"`
	RecordTTL   int    `json:"record_ttl,omitempty"`
}
