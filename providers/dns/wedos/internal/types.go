package internal

import "encoding/json"

const codeOk = 1000

const (
	commandPing            = "ping"
	commandDNSDomainCommit = "dns-domain-commit"
	commandDNSRowsList     = "dns-rows-list"
	commandDNSRowDelete    = "dns-row-delete"
	commandDNSRowAdd       = "dns-row-add"
	commandDNSRowUpdate    = "dns-row-update"
)

type Response interface {
	GetCode() int
	GetResult() string
}

type APIResponse[D any] struct {
	Response ResponsePayload[D] `json:"response"`
}

func (a APIResponse[D]) GetCode() int {
	return a.Response.Code
}

func (a APIResponse[D]) GetResult() string {
	return a.Response.Result
}

type ResponsePayload[D any] struct {
	Code      int    `json:"code,omitempty"`
	Result    string `json:"result,omitempty"`
	Timestamp int    `json:"timestamp,omitempty"`
	SvTRID    string `json:"svTRID,omitempty"`
	Command   string `json:"command,omitempty"`
	Data      D      `json:"data"`
}

type Rows struct {
	Rows []DNSRow `json:"row"`
}

type DNSRow struct {
	ID   string      `json:"ID,omitempty"`
	Name string      `json:"name,omitempty"`
	TTL  json.Number `json:"ttl,omitempty"`
	Type string      `json:"rdtype,omitempty"`
	Data string      `json:"rdata"`
}

type DNSRowRequest struct {
	ID     string      `json:"row_id,omitempty"`
	Domain string      `json:"domain,omitempty"`
	Name   string      `json:"name,omitempty"`
	TTL    json.Number `json:"ttl,omitempty"`
	Type   string      `json:"type,omitempty"`
	Data   string      `json:"rdata"`
}

type APIRequest struct {
	User    string `json:"user,omitempty"`
	Auth    string `json:"auth,omitempty"`
	Command string `json:"command,omitempty"`
	Data    any    `json:"data,omitempty"`
}
