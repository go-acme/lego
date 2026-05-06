package internal

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	APIResponse
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Code, a.Message)
}

type APIResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

type APIValue struct {
	Value string `json:"value"`
}

type Session struct {
	ID APIValue `json:"sess_id"`
}

type LoginRequest struct {
	Email      string `url:"email,omitempty"`
	Password   string `url:"password,omitempty"`
	OrderID    string `url:"ord_id,omitempty"`
	APIVersion string `url:"api_version,omitempty"`
}

type GetRecordsRequest struct {
	Page         string `url:"dns_records_load_page,omitempty"`
	OnlyForDomID string `url:"dns_records_load_only_for_dom_id,omitempty"`
	Keyword      string `url:"dns_records_load_keyword,omitempty"`
	Subdomain    string `url:"dns_records_load_subdomain,omitempty"`
	Content      string `url:"dns_records_load_content,omitempty"`
	Type         string `url:"dns_records_load_type,omitempty"`
}

type Records struct {
	CountMax   APIValue `json:"dns_record_count_max"`
	CountTotal APIValue `json:"dns_record_count_total"`
	LoadPage   APIValue `json:"dns_records_load_page"`
	Domains    []Domain `json:"domains"`
}

type Domain struct {
	ID                      APIValue    `json:"dom_id"`
	Domain                  APIValue    `json:"dom_domain"`
	NameserversUsingDefault APIValue    `json:"nameservers_using_default"`
	DNSRecordCount          APIValue    `json:"dns_record_count"`
	DNSRecords              []DNSRecord `json:"dns_records"`
}

type DNSRecord struct {
	ID        APIValue `json:"id"`
	Name      APIValue `json:"name"`
	Type      APIValue `json:"type"`
	Subdomain APIValue `json:"subdomain"`
	Content   APIValue `json:"content"`
	TTL       APIValue `json:"ttl"`
	Priority  APIValue `json:"prio"`
}

type SetRecordRequest struct {
	DomainID  string `url:"dom_id,omitempty"`
	ID        string `url:"dns_record_id,omitempty"`
	Subdomain string `url:"subdomain,omitempty"`
	Type      string `url:"type,omitempty"`
	Content   string `url:"content,omitempty"`
	TTL       int    `url:"ttl,omitempty"`
	Priority  int    `url:"prio,omitempty"`
}
