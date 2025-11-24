package internal

import (
	"fmt"
	"time"
)

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
	ID               string `json:"zone_id,omitempty"`
	Name             string `json:"zone_name,omitempty"`
	NameDisplay      string `json:"zone_name_display,omitempty"`
	Type             string `json:"zone_type,omitempty"`
	Active           string `json:"zone_active,omitempty"`
	Protected        string `json:"zone_protected,omitempty"`
	IsRegistered     string `json:"zone_is_registered,omitempty"`
	Updated          int    `json:"zone_updated,omitempty"`
	CustomerID       string `json:"cust_id,omitempty"`
	DomainRegistrar  string `json:"domain_registrar,omitempty"`
	DomainStatus     string `json:"domain_status,omitempty"`
	DomainExpiryDate string `json:"domain_expiry_date,omitempty"`
	DomainAutoRenew  string `json:"domain_auto_renew,omitempty"`
	ExternalDNS      string `json:"external_dns,omitempty"`
	RecordCount      int    `json:"record_count,omitempty"`
}

type Record struct {
	ID    string `json:"record_id,omitempty"`
	Name  string `json:"record_name,omitempty"`
	Type  string `json:"record_type,omitempty"`
	Value string `json:"record_value,omitempty"`
	TTL   int    `json:"record_ttl,omitempty"`
}

type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Code     int    `json:"code,omitempty"`
}

type Token struct {
	Token       string `json:"token"`
	TokenExpire int64  `json:"token_expire"`
	CustomerID  string `json:"customer_id"`
}

func (t *Token) IsExpired() bool {
	if t == nil {
		return true
	}

	return time.Now().UTC().Add(1 * time.Minute).After(time.Unix(t.TokenExpire, 0).UTC())
}
