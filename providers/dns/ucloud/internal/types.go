package internal

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

type DomainDNSAddRequest struct {
	request.CommonBase

	Domain     *string `json:"Dn,omitempty"`
	RecordName *string `json:"RecordName,omitempty"`
	Type       *string `json:"DnsType,omitempty"`
	Content    *string `json:"Content,omitempty"`
	TTL        *string `json:"TTL,omitempty"`
	Priority   *string `json:"Prio,omitempty"`
}

type DomainDNSAddResponse struct {
	response.CommonBase
}

type DomainDNSQueryRequest struct {
	request.CommonBase

	Domain *string `json:"Dn,omitempty"`
}

type DomainDNSQueryResponse struct {
	response.CommonBase

	Data []DomainDNSRecord
}

type DomainDNSRecord struct {
	Type     string `json:"DnsType,omitempty"`
	Name     string `json:"RecordName,omitempty"`
	Content  string `json:"Content,omitempty"`
	Priority string `json:"Prio,omitempty"`
	TTL      string `json:"TTL,omitempty"`
}

type DeleteDNSRecordRequest struct {
	request.CommonBase

	Domain     *string `json:"Dn,omitempty"`
	RecordName *string `json:"RecordName,omitempty"`
	Type       *string `json:"DnsType,omitempty"`
	Content    *string `json:"Content,omitempty"`
}

type DeleteDNSRecordResponse struct {
	response.CommonBase
}
