package internal

import (
	"fmt"
	"strings"
)

type Record struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
	Type    string `json:"type,omitempty"`
	Comment string `json:"comment,omitempty"`
	Content string `json:"content,omitempty"`
}

type APIResponse[T any] struct {
	Errors     Errors      `json:"errors,omitempty"`
	Messages   []Message   `json:"messages,omitempty"`
	Success    bool        `json:"success,omitempty"`
	Result     T           `json:"result,omitempty"`
	ResultInfo *ResultInfo `json:"result_info,omitempty"`
}

type Message struct {
	Code             int          `json:"code"`
	Message          string       `json:"message"`
	DocumentationURL string       `json:"documentation_url"`
	Source           *Source      `json:"source"`
	ErrorChain       []ErrorChain `json:"error_chain"`
}

type Source struct {
	Pointer string `json:"pointer"`
}

type ErrorChain struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Errors []Message

func (e Errors) Error() string {
	var msg strings.Builder

	for _, item := range e {
		msg.WriteString(fmt.Sprintf("%d: %s", item.Code, item.Message))

		for _, link := range item.ErrorChain {
			msg.WriteString(fmt.Sprintf("; %d: %s", link.Code, link.Message))
		}
	}

	return msg.String()
}

type ResultInfo struct {
	Count      int `json:"count"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

type Zone struct {
	ID                string     `json:"id"`
	Account           Account    `json:"account"`
	Meta              Meta       `json:"meta"`
	Name              string     `json:"name"`
	Owner             Owner      `json:"owner"`
	Plan              Plan       `json:"plan"`
	CnameSuffix       string     `json:"cname_suffix"`
	Paused            bool       `json:"paused"`
	Permissions       []string   `json:"permissions"`
	Tenant            Tenant     `json:"tenant"`
	TenantUnit        TenantUnit `json:"tenant_unit"`
	Type              string     `json:"type"`
	VanityNameServers []string   `json:"vanity_name_servers"`
}

type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Meta struct {
	CdnOnly                bool `json:"cdn_only"`
	CustomCertificateQuota int  `json:"custom_certificate_quota"`
	DNSOnly                bool `json:"dns_only"`
	FoundationDNS          bool `json:"foundation_dns"`
	PageRuleQuota          int  `json:"page_rule_quota"`
	PhishingDetected       bool `json:"phishing_detected"`
	Step                   int  `json:"step"`
}

type Owner struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Plan struct {
	ID                string `json:"id"`
	CanSubscribe      bool   `json:"can_subscribe"`
	Currency          string `json:"currency"`
	ExternallyManaged bool   `json:"externally_managed"`
	Frequency         string `json:"frequency"`
	IsSubscribed      bool   `json:"is_subscribed"`
	LegacyDiscount    bool   `json:"legacy_discount"`
	LegacyID          string `json:"legacy_id"`
	Price             int    `json:"price"`
	Name              string `json:"name"`
}

type Tenant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TenantUnit struct {
	ID string `json:"id"`
}
