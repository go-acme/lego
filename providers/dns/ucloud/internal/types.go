package internal

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

type DomainDNSAddRequest struct {
	request.CommonBase

	Dn         *string
	RecordName *string
	DnsType    *string //nolint:revive // Because the struct names are used directly.
	Content    *string
	TTL        *string
	Prio       *string
}

type DomainDNSAddResponse struct {
	response.CommonBase
}

type DomainDNSQueryRequest struct {
	request.CommonBase

	Dn *string
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

	Dn         *string
	RecordName *string
	DnsType    *string //nolint:revive // Because the struct names are used directly.
	Content    *string
}

type DeleteDNSRecordResponse struct {
	response.CommonBase
}
