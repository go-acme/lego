package internal

import (
	"fmt"
)

type APIError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Code, a.Message)
}

type Domain struct {
	ID                 int    `json:"id,omitempty"`
	UserID             int    `json:"id_user,omitempty"`
	ParentDomainID     int    `json:"id_parent_domain,omitempty"`
	Status             int    `json:"status,omitempty"`
	Domain             string `json:"domain,omitempty"`
	DomainUTF8         string `json:"domain_utf8,omitempty"`
	IsDisabled         bool   `json:"is_disabled,omitempty"`
	IsCustomDNS        bool   `json:"is_custom_dns,omitempty"`
	IsDNSDisabled      bool   `json:"is_dns_disabled,omitempty"`
	IsSubdomain        bool   `json:"is_subdomain,omitempty"`
	IsSystemDomain     bool   `json:"is_system_domain,omitempty"`
	IsEmailDomain      bool   `json:"is_email_domain,omitempty"`
	IsEmailSendingOnly bool   `json:"is_email_sending_only,omitempty"`
}

type DomainID struct {
	ID int `json:"id,omitempty"`
}

type DomainRecords struct {
	IsCustomDNS   bool     `json:"is_custom_dns,omitempty"`
	IsDNSDisabled bool     `json:"is_dns_disabled,omitempty"`
	DkimRecord    string   `json:"dkim_record,omitempty"`
	Records       *Records `json:"records,omitempty"`
}

type Records struct {
	Soa   *SOARecord `json:"soa,omitempty"`
	Other []Record   `json:"other,omitempty"`
}

type SOARecord struct {
	TTL       int    `json:"ttl,omitempty"`
	PrimaryNs string `json:"primary_ns,omitempty"`
	RName     string `json:"rname,omitempty"`
	Refresh   int    `json:"refresh,omitempty"`
	Retry     int    `json:"retry,omitempty"`
	Expire    int    `json:"expire,omitempty"`
	Minimum   int    `json:"minimum,omitempty"`
}

type Record struct {
	Host  string `json:"host"`
	TTL   int    `json:"ttl"`
	Type  string `json:"type"`
	Value string `json:"value"`
}
