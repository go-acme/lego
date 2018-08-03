// Package netcup implements a DNS Provider for solving the DNS-01 challenge using the netcup DNS API.
package netcup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// NetcupBaseUrl for reaching the jSON-based API-Endpoint of netcup
const netcupBaseUrl = "https://ccp.netcup.net/run/webservice/servers/endpoint.php?JSON"

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	customernumber string
	apiKey         string
	apipassword    string
	client         http.Client
}

// Request wrapper as specified in netcup wiki
// needed for every request to netcup API around *Msg
// https://www.netcup-wiki.de/wiki/CCP_API#Anmerkungen_zu_JSON-Requests
type Request struct {
	Action string      `json:"action"`
	Param  interface{} `json:"param"`
}

// LoginMsg as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#login
type LoginMsg struct {
	Customernumber  string `json:"customernumber"`
	Apikey          string `json:"apikey"`
	Apipassword     string `json:"apipassword"`
	Clientrequestid string `json:"clientrequestid,omitempty"`
}

// LogoutMsg as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#logout
type LogoutMsg struct {
	Customernumber  string `json:"customernumber"`
	Apikey          string `json:"apikey"`
	Apisessionid    string `json:"apisessionid"`
	Clientrequestid string `json:"clientrequestid,omitempty"`
}

// UpdateDNSRecordsMsg as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#updateDnsRecords
type UpdateDNSRecordsMsg struct {
	Domainname      string       `json:"domainname"`
	Customernumber  string       `json:"customernumber"`
	Apikey          string       `json:"apikey"`
	Apisessionid    string       `json:"apisessionid"`
	Clientrequestid string       `json:"clientrequestid,omitempty"`
	Dnsrecordset    DNSRecordSet `json:"dnsrecordset"`
}

// DNSRecordSet as specified in netcup WSDL
// needed in UpdateDNSRecordsMsg
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#Dnsrecordset
type DNSRecordSet struct {
	Dnsrecords []DNSRecord `json:"dnsrecords"`
}

// InfoDnsRecordsMsg as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#infoDnsRecords
type InfoDNSRecordsMsg struct {
	Domainname      string `json:"domainname"`
	Customernumber  string `json:"customernumber"`
	Apikey          string `json:"apikey"`
	Apisessionid    string `json:"apisessionid"`
	Clientrequestid string `json:"clientrequestid,omitempty"`
}

// DNSRecord as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#Dnsrecord
type DNSRecord struct {
	Id           int    `json:"id,string,omitempty"`
	Hostname     string `json:"hostname"`
	Recordtype   string `json:"type"`
	Priority     string `json:"priority,omitempty"`
	Destination  string `json:"destination"`
	Deleterecord bool   `json:"deleterecord,omitempty"`
	State        string `json:"state,omitempty"`
}

// ResponseMsg as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#Responsemessage
type ResponseMsg struct {
	Serverrequestid string       `json:"serverrequestid"`
	Clientrequestid string       `json:"clientrequestid,omitempty"`
	Action          string       `json:"action"`
	Status          string       `json:"status"`
	Statuscode      int          `json:"statuscode"`
	Shortmessage    string       `json:"shortmessage"`
	Longmessage     string       `json:"longmessage"`
	Responsedata    ResponseData `json:"responsedata,omitempty"`
}

// LogoutResponseMsg similar to ResponseMsg
// allows empty ResponseData field whilst unmarshaling
type LogoutResponseMsg struct {
	Serverrequestid string `json:"serverrequestid"`
	Clientrequestid string `json:"clientrequestid,omitempty"`
	Action          string `json:"action"`
	Status          string `json:"status"`
	Statuscode      int    `json:"statuscode"`
	Shortmessage    string `json:"shortmessage"`
	Longmessage     string `json:"longmessage"`
	Responsedata    string `json:"responsedata,omitempty"`
}

// ResponseData to enable correct unmarshaling of ResponseMsg
type ResponseData struct {
	Apisessionid string      `json:"apisessionid"`
	Dnsrecords   []DNSRecord `json:"dnsrecords"`
}

// NewDNSProvider returns a DNSProvider instance configured for netcup.
// Credentials must be passed in the environment variables: NETCUP_CUSTOMER_NUMBER,
// NETCUP_API_KEY, NETCUP_API_PASSWORD
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("NETCUP_CUSTOMER_NUMBER", "NETCUP_API_KEY", "NETCUP_API_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("netcup: %v", err)
	}

	return NewDNSProviderCredentials(values["NETCUP_CUSTOMER_NUMBER"], values["NETCUP_API_KEY"], values["NETCUP_API_PASSWORD"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for netcup.
func NewDNSProviderCredentials(customer, key, password string) (*DNSProvider, error) {
	if customer == "" || key == "" || password == "" {
		return nil, fmt.Errorf("netcup: netcup credentials missing!")
	}

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	return &DNSProvider{
		customernumber: customer,
		apiKey:         key,
		apipassword:    password,
		client:         client,
	}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (d *DNSProvider) Present(domainname, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domainname, keyAuth)

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("netcup: failed to find DNSZone, %v", err)
	}

	hostname := strings.Replace(fqdn, "."+zone, "", 1)

	zone = acme.UnFqdn(zone)

	record := createTxtRecord(hostname, value)

	sessionid, err := d.login()
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	err = d.updateDnsRecord(sessionid, zone, record)
	if err != nil {
		err2 := d.logout(sessionid)
		if err2 != nil {
			return fmt.Errorf("failed to add TXT-Record: %v \n %v", err, err2)
		}
		return fmt.Errorf("failed to add TXT-Record: %v", err)
	}

	err = d.logout(sessionid)
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domainname, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domainname, keyAuth)

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("netcup: failed to find DNSZone, %v", err)
	}

	hostname := strings.Replace(fqdn, "."+zone, "", 1)

	zone = acme.UnFqdn(zone)

	record := createTxtRecord(hostname, value)

	sessionid, err := d.login()
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	records, err := d.getDnsRecords(zone, sessionid)
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	idx, err := d.getDnsRecordIdx(records, record)
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	records[idx].Deleterecord = true

	err = d.updateDnsRecord(sessionid, zone, records[idx])
	if err != nil {
		err2 := d.logout(sessionid)
		if err2 != nil {
			return fmt.Errorf("netcup: " + err.Error() + "\n netcup: " + err2.Error())
		}
		return fmt.Errorf("netcup: %v", err)
	}

	err = d.logout(sessionid)
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}
	return nil
}

// createTxtRecord uses the supplied values to return a
// DNSRecord of type TXT for the dns-01 challenge
func createTxtRecord(hostname, value string) DNSRecord {
	return DNSRecord{
		Id:           0,
		Hostname:     hostname,
		Recordtype:   "TXT",
		Priority:     "",
		Destination:  value,
		Deleterecord: false,
		State:        "",
	}

}

// login performs the login as specified by the netcup WSDL
// returns sessionid needed to perform remaining actions
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (d *DNSProvider) login() (string, error) {
	msg := &LoginMsg{
		Customernumber:  d.customernumber,
		Apikey:          d.apiKey,
		Apipassword:     d.apipassword,
		Clientrequestid: "",
	}

	payload := &Request{
		Action: "login",
		Param:  msg,
	}

	response, err := d.sendRequest(payload)
	if err != nil {
		return "", fmt.Errorf("error sending request to DNS-API, %v", err)
	}

	var r ResponseMsg

	err = json.Unmarshal(response, &r)
	if err != nil {
		return "", fmt.Errorf("error decoding response of DNS-API, %v", err)
	}
	if r.Status != "success" {
		return "", fmt.Errorf("error logging into DNS-API, %v", r.Longmessage)
	}
	return r.Responsedata.Apisessionid, nil
}

// logout performs the logout with the supplied sessionid as specified by the netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (d *DNSProvider) logout(sessionid string) error {
	msg := &LogoutMsg{
		Customernumber:  d.customernumber,
		Apikey:          d.apiKey,
		Apisessionid:    sessionid,
		Clientrequestid: "",
	}

	payload := &Request{
		Action: "logout",
		Param:  msg,
	}

	response, err := d.sendRequest(payload)
	if err != nil {
		return fmt.Errorf("Error logging out of DNS-API: %v", err)
	}

	var r LogoutResponseMsg

	err = json.Unmarshal(response, &r)
	if err != nil {
		return fmt.Errorf("Error logging out of DNS-API: %v", err)
	}

	if r.Status != "success" {
		return fmt.Errorf("Error logging out of DNS-API: %v", r.Shortmessage)
	}
	return nil
}

// updateDnsRecord performs an update of the DNSRecords as specified by the netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (d *DNSProvider) updateDnsRecord(sessionid, domainname string, record DNSRecord) error {
	msg := UpdateDNSRecordsMsg{
		Domainname:      domainname,
		Customernumber:  d.customernumber,
		Apikey:          d.apiKey,
		Apisessionid:    sessionid,
		Clientrequestid: "",
		Dnsrecordset:    DNSRecordSet{[]DNSRecord{record}},
	}

	payload := Request{
		Action: "updateDnsRecords",
		Param:  msg,
	}

	response, err := d.sendRequest(payload)
	if err != nil {
		return err
	}

	var r ResponseMsg

	err = json.Unmarshal(response, &r)
	if err != nil {
		return err
	}

	if r.Status != "success" {
		return fmt.Errorf("ServerrequestId: %v, Status: %v, Statuscode: %v, Message: %v", r.Serverrequestid, r.Status, r.Statuscode, r.Longmessage)
	}
	return nil
}

// getDnsRecords retrieves all dns records of an DNS-Zone as specified by the netcup WSDL
// returns an array of DNSRecords
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (d *DNSProvider) getDnsRecords(hostname, apisessionid string) ([]DNSRecord, error) {
	msg := InfoDNSRecordsMsg{
		Domainname:      hostname,
		Customernumber:  d.customernumber,
		Apikey:          d.apiKey,
		Apisessionid:    apisessionid,
		Clientrequestid: "",
	}
	payload := Request{
		Action: "infoDnsRecords",
		Param:  msg,
	}

	response, err := d.sendRequest(payload)
	if err != nil {
		return nil, err
	}

	var r ResponseMsg

	err = json.Unmarshal(response, &r)
	if err != nil {
		return nil, err
	}

	if r.Status != "success" {
		return nil, fmt.Errorf(r.Shortmessage)
	}
	return r.Responsedata.Dnsrecords, nil

}

// getDnsRecordIdx searches a given array of DNSRecords for a given DNSRecord
// equivalence is determined by Destination and RecortType attributes
// returns index of given DNSRecord in given array of DNSRecords
func (d *DNSProvider) getDnsRecordIdx(records []DNSRecord, record DNSRecord) (int, error) {
	for index, element := range records {
		if record.Destination == element.Destination && record.Recordtype == element.Recordtype {
			return index, nil
		}
	}
	return -1, fmt.Errorf("No DNS Record found")
}

// sendRequest marshals given body to JSON, send the request to netcup API
// and returns body of response
func (d *DNSProvider) sendRequest(payload interface{}) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, netcupBaseUrl, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Close = true

	//req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("User-Agent", "LEGO_ACME_DNS_Client")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("API request failed with HTTP Status code %d", resp.StatusCode)
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read of response body failed, %v", err)
	}
	defer resp.Body.Close()

	return body, nil
}
