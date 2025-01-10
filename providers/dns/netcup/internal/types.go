package internal

import (
	"encoding/json"
	"fmt"
)

// success response status.
const success = "success"

// Request wrapper as specified in netcup wiki
// needed for every request to netcup API around *Msg.
// https://www.netcup-wiki.de/wiki/CCP_API#Anmerkungen_zu_JSON-Requests
type Request struct {
	Action string `json:"action"`
	Param  any    `json:"param"`
}

// LoginRequest as specified in netcup WSDL.
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#login
type LoginRequest struct {
	CustomerNumber  string `json:"customernumber"`
	APIKey          string `json:"apikey"`
	APIPassword     string `json:"apipassword"`
	ClientRequestID string `json:"clientrequestid,omitempty"`
}

// LogoutRequest as specified in netcup WSDL.
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#logout
type LogoutRequest struct {
	CustomerNumber  string `json:"customernumber"`
	APIKey          string `json:"apikey"`
	APISessionID    string `json:"apisessionid"`
	ClientRequestID string `json:"clientrequestid,omitempty"`
}

// UpdateDNSRecordsRequest as specified in netcup WSDL.
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#updateDnsRecords
type UpdateDNSRecordsRequest struct {
	DomainName      string       `json:"domainname"`
	CustomerNumber  string       `json:"customernumber"`
	APIKey          string       `json:"apikey"`
	APISessionID    string       `json:"apisessionid"`
	ClientRequestID string       `json:"clientrequestid,omitempty"`
	DNSRecordSet    DNSRecordSet `json:"dnsrecordset"`
}

// DNSRecordSet as specified in netcup WSDL.
// needed in UpdateDNSRecordsRequest.
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#Dnsrecordset
type DNSRecordSet struct {
	DNSRecords []DNSRecord `json:"dnsrecords"`
}

// InfoDNSRecordsRequest as specified in netcup WSDL.
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#infoDnsRecords
type InfoDNSRecordsRequest struct {
	DomainName      string `json:"domainname"`
	CustomerNumber  string `json:"customernumber"`
	APIKey          string `json:"apikey"`
	APISessionID    string `json:"apisessionid"`
	ClientRequestID string `json:"clientrequestid,omitempty"`
}

// DNSRecord as specified in netcup WSDL.
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#Dnsrecord
type DNSRecord struct {
	ID           int    `json:"id,string,omitempty"`
	Hostname     string `json:"hostname"`
	RecordType   string `json:"type"`
	Priority     string `json:"priority,omitempty"`
	Destination  string `json:"destination"`
	DeleteRecord bool   `json:"deleterecord,omitempty"`
	State        string `json:"state,omitempty"`
}

// ResponseMsg as specified in netcup WSDL.
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#Responsemessage
type ResponseMsg struct {
	ServerRequestID string          `json:"serverrequestid"`
	ClientRequestID string          `json:"clientrequestid,omitempty"`
	Action          string          `json:"action"`
	Status          string          `json:"status"`
	StatusCode      int             `json:"statuscode"`
	ShortMessage    string          `json:"shortmessage"`
	LongMessage     string          `json:"longmessage"`
	ResponseData    json.RawMessage `json:"responsedata,omitempty"`
}

func (r *ResponseMsg) Error() string {
	return fmt.Sprintf("an error occurred during the action %s: [Status=%s, StatusCode=%d, ShortMessage=%s, LongMessage=%s]",
		r.Action, r.Status, r.StatusCode, r.ShortMessage, r.LongMessage)
}

// LoginResponse response to login action.
type LoginResponse struct {
	APISessionID string `json:"apisessionid"`
}

// InfoDNSRecordsResponse response to infoDnsRecords action.
type InfoDNSRecordsResponse struct {
	APISessionID string      `json:"apisessionid"`
	DNSRecords   []DNSRecord `json:"dnsrecords,omitempty"`
}
