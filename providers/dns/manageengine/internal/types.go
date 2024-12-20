package internal

import (
	"strings"
)

type APIError struct {
	Message string `json:"error"`
	Detail  string `json:"detail"`
}

func (a *APIError) Error() string {
	var msg []string

	if a.Message != "" {
		msg = append(msg, a.Message)
	}

	if a.Detail != "" {
		msg = append(msg, a.Detail)
	}

	return strings.Join(msg, " ")
}

type APIResponse struct {
	Message string `json:"message,omitempty"`
}

type ZoneRecord struct {
	ZoneID           int      `json:"zone_id,omitempty"`
	SpfTxtDomainID   int      `json:"spf_txt_domain_id,omitempty"`
	DomainName       string   `json:"domain_name,omitempty"`
	DomainTTL        int      `json:"domain_ttl,omitempty"`
	DomainLocationID int      `json:"domain_location_id,omitempty"`
	RecordType       string   `json:"record_type,omitempty"`
	Records          []Record `json:"records"`
}

type Record struct {
	ID       int      `json:"record_id,omitempty"`
	Values   []string `json:"value,omitempty"`
	Disabled bool     `json:"disabled,omitempty"`
	DomainID int      `json:"domain_id,omitempty"`
}

type Zone struct {
	ZoneID        int      `json:"zone_id"`
	ZoneName      string   `json:"zone_name"`
	ZoneTTL       int      `json:"zone_ttl"`
	ZoneType      int      `json:"zone_type,omitempty"`
	ZoneTargeting bool     `json:"zone_targeting"`
	Refresh       int      `json:"refresh"`
	Retry         int      `json:"retry"`
	Expiry        int      `json:"expiry"`
	Minimum       int      `json:"minimum"`
	Org           int      `json:"org"`
	AnyQuery      bool     `json:"any_query"`
	Vanity        bool     `json:"vanity,omitempty"`
	NsID          int      `json:"ns_id"`
	Serial        int      `json:"serial"`
	Nss           []string `json:"ns"`
}
