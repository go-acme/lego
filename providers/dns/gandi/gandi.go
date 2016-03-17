// Package gandi implements a DNS provider for solving the DNS-01
// challenge using Gandi DNS.
package gandi

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/xenolf/lego/acme"
)

// Gandi API reference:       http://doc.rpc.gandi.net/index.html
// Gandi API domain examples: http://doc.rpc.gandi.net/domain/faq.html

type inProgressInfo struct {
	zoneID    int    // zoneID of zone to restore in CleanUp
	newZoneID int    // zoneID of temporary zone containing TXT record
	rootDN    string // the registered (root) domain name being manipulated
}

// DNSProvider is an implementation of the
// acme.ChallengeProviderTimeout interface that uses Gandi's XML-RPC
// API to manage TXT records for a domain.
type DNSProvider struct {
	apiKey            string
	inProgressFQDNs   map[string]inProgressInfo
	inProgressRootDNs map[string]struct{}
	inProgressMu      sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Gandi.
// Credentials must be passed in the environment variable: GANDI_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	apiKey := os.Getenv("GANDI_API_KEY")
	return NewDNSProviderCredentials(apiKey)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for Gandi.
func NewDNSProviderCredentials(apiKey string) (*DNSProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("No Gandi API Key given")
	}
	return &DNSProvider{
		apiKey:            apiKey,
		inProgressFQDNs:   make(map[string]inProgressInfo),
		inProgressRootDNs: make(map[string]struct{}),
	}, nil
}

// Present creates a TXT record using the specified parameters. It
// does this by creating and activating a new temporary DNS zone. This
// new zone contains the TXT record.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	if ttl < 300 {
		ttl = 300 // 300 is gandi minimum value for ttl
	}
	i := strings.Index(fqdn, ".")
	sub := fqdn[:i+1]
	root := fqdn[i+1:]
	var zoneID int
	var err error
	// find sub and root (sub + root == fqdn) where root is the domain
	// registered with gandi. Do this by successively increasing sub
	// and decreasing root until root matches a registered domain with
	// a zone_id
	for {
		zoneID, err = d.getZoneID(root)
		if err == nil {
			// domain found
			break
		}
		if faultErr, ok := err.(rpcError); ok {
			if faultErr.faultCode == 510042 {
				// 510042 error means root is not found - increase
				// sub, reduce root and retry.
				// [see http://doc.rpc.gandi.net/errors/fault_codes.html]
				i := strings.Index(root, ".")
				if i != -1 && i != len(root)-1 &&
					strings.Index(root[i+1:], ".") != -1 &&
					strings.Index(root[i+1:], ".") != len(root[i+1:])-1 {
					sub = sub + root[:i+1]
					root = root[i+1:]
					continue
				}
			}
		}
		// root is not found and cannot be reduced in size any further
		// or there is some other error from getZoneID
		return err
	}
	// remove trailing "." from sub
	sub = sub[:len(sub)-1]
	// acquire lock and check there is not a challenge already in
	// progress for this value of root
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()
	if _, ok := d.inProgressRootDNs[root]; ok {
		return fmt.Errorf(
			"Gandi DNS: challenge already in progress on root domain")
	}
	// perform API actions to create and activate new zone for root
	// containing the required TXT record
	newZoneName := fmt.Sprintf(
		"%s [ACME Challenge %s]",
		root[:len(root)-1], time.Now().Format(time.RFC822Z))
	newZoneID, err := d.cloneZone(zoneID, newZoneName)
	if err != nil {
		return err
	}
	newZoneVersion, err := d.newZoneVersion(newZoneID)
	if err != nil {
		return err
	}
	err = d.addTXTRecord(newZoneID, newZoneVersion, sub, value, ttl)
	if err != nil {
		return err
	}
	err = d.setZoneVersion(newZoneID, newZoneVersion)
	if err != nil {
		return err
	}
	err = d.setZone(root, newZoneID)
	if err != nil {
		return err
	}
	// save data necessary for CleanUp
	d.inProgressFQDNs[fqdn] = inProgressInfo{
		zoneID:    zoneID,
		newZoneID: newZoneID,
		rootDN:    root,
	}
	d.inProgressRootDNs[root] = struct{}{}
	return nil
}

// CleanUp removes the TXT record matching the specified
// parameters. It does this by restoring the old DNS zone and removing
// the temporary one created by Present.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)
	// acquire lock and retrieve zoneID, newZoneID and root
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()
	if _, ok := d.inProgressFQDNs[fqdn]; !ok {
		// if there is no cleanup information then just return
		return nil
	}
	zoneID := d.inProgressFQDNs[fqdn].zoneID
	newZoneID := d.inProgressFQDNs[fqdn].newZoneID
	root := d.inProgressFQDNs[fqdn].rootDN
	delete(d.inProgressFQDNs, fqdn)
	delete(d.inProgressRootDNs, root)
	// perform API actions to restore old zone for root
	err := d.setZone(root, zoneID)
	if err != nil {
		return err
	}
	err = d.deleteZone(newZoneID)
	if err != nil {
		return err
	}
	return nil
}

// Timeout returns the values (40*time.Minute, 60*time.Second) which
// are used by the acme package as timeout and check interval values
// when checking for DNS record propagation with Gandi.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 40 * time.Minute, 60 * time.Second
}

// Endpoint is the Gandi XML-RPC endpoint used by Present and
// CleanUp. It is exported only so it may be overridden during package
// tests.
var Endpoint = "https://rpc.gandi.net/xmlrpc/"

// types for XML-RPC method calls and parameters

type param interface {
	param()
}
type paramString struct {
	XMLName xml.Name `xml:"param"`
	Value   string   `xml:"value>string"`
}
type paramInt struct {
	XMLName xml.Name `xml:"param"`
	Value   int      `xml:"value>int"`
}

type structMember interface {
	structMember()
}
type structMemberString struct {
	Name  string `xml:"name"`
	Value string `xml:"value>string"`
}
type structMemberInt struct {
	Name  string `xml:"name"`
	Value int    `xml:"value>int"`
}
type paramStruct struct {
	XMLName       xml.Name       `xml:"param"`
	StructMembers []structMember `xml:"value>struct>member"`
}

func (p paramString) param()               {}
func (p paramInt) param()                  {}
func (m structMemberString) structMember() {}
func (m structMemberInt) structMember()    {}
func (p paramStruct) param()               {}

type methodCall struct {
	XMLName    xml.Name `xml:"methodCall"`
	MethodName string   `xml:"methodName"`
	Params     []param  `xml:"params"`
}

// types for XML-RPC responses

type response interface {
	faultCode() int
	faultString() string
}

type responseFault struct {
	FaultCode   int    `xml:"fault>value>struct>member>value>int"`
	FaultString string `xml:"fault>value>struct>member>value>string"`
}

func (r responseFault) faultCode() int      { return r.FaultCode }
func (r responseFault) faultString() string { return r.FaultString }

type responseStruct struct {
	responseFault
	StructMembers []struct {
		Name     string `xml:"name"`
		ValueInt int    `xml:"value>int"`
	} `xml:"params>param>value>struct>member"`
}

type responseInt struct {
	responseFault
	Value int `xml:"params>param>value>int"`
}

type responseBool struct {
	responseFault
	Value bool `xml:"params>param>value>boolean"`
}

// POSTing/Marshalling/Unmarshalling

type rpcError struct {
	faultCode   int
	faultString string
}

func (e rpcError) Error() string {
	return fmt.Sprintf(
		"Gandi DNS: RPC Error: (%d) %s", e.faultCode, e.faultString)
}

func httpPost(url string, bodyType string, body io.Reader) ([]byte, error) {
	client := http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(url, bodyType, body)
	if err != nil {
		return nil, fmt.Errorf("Gandi DNS: HTTP Post Error: %v", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Gandi DNS: HTTP Post Error: %v", err)
	}
	return b, nil
}

// rpcCall makes an XML-RPC call to Gandi's RPC endpoint by
// marshalling the data given in the call argument to XML and sending
// that via HTTP Post to Gandi. The response is then unmarshalled into
// the resp argument.
func rpcCall(call *methodCall, resp response) error {
	// marshal
	b, err := xml.MarshalIndent(call, "", "  ")
	if err != nil {
		return fmt.Errorf("Gandi DNS: Marshal Error: %v", err)
	}
	// post
	b = append([]byte(`<?xml version="1.0"?>`+"\n"), b...)
	respBody, err := httpPost(Endpoint, "text/xml", bytes.NewReader(b))
	if err != nil {
		return err
	}
	// unmarshal
	err = xml.Unmarshal(respBody, resp)
	if err != nil {
		return fmt.Errorf("Gandi DNS: Unmarshal Error: %v", err)
	}
	if resp.faultCode() != 0 {
		return rpcError{
			faultCode: resp.faultCode(), faultString: resp.faultString()}
	}
	return nil
}

// functions to perform API actions

func (d *DNSProvider) getZoneID(domain string) (int, error) {
	resp := &responseStruct{}
	err := rpcCall(&methodCall{
		MethodName: "domain.info",
		Params: []param{
			paramString{Value: d.apiKey},
			paramString{Value: domain},
		},
	}, resp)
	if err != nil {
		return 0, err
	}
	var zoneID int
	for _, member := range resp.StructMembers {
		if member.Name == "zone_id" {
			zoneID = member.ValueInt
		}
	}
	if zoneID == 0 {
		return 0, fmt.Errorf("Gandi DNS: Could not determine zone_id")
	}
	return zoneID, nil
}

func (d *DNSProvider) cloneZone(zoneID int, name string) (int, error) {
	resp := &responseStruct{}
	err := rpcCall(&methodCall{
		MethodName: "domain.zone.clone",
		Params: []param{
			paramString{Value: d.apiKey},
			paramInt{Value: zoneID},
			paramInt{Value: 0},
			paramStruct{
				StructMembers: []structMember{
					structMemberString{
						Name:  "name",
						Value: name,
					}},
			},
		},
	}, resp)
	if err != nil {
		return 0, err
	}
	var newZoneID int
	for _, member := range resp.StructMembers {
		if member.Name == "id" {
			newZoneID = member.ValueInt
		}
	}
	if newZoneID == 0 {
		return 0, fmt.Errorf("Gandi DNS: Could not determine cloned zone_id")
	}
	return newZoneID, nil
}

func (d *DNSProvider) newZoneVersion(zoneID int) (int, error) {
	resp := &responseInt{}
	err := rpcCall(&methodCall{
		MethodName: "domain.zone.version.new",
		Params: []param{
			paramString{Value: d.apiKey},
			paramInt{Value: zoneID},
		},
	}, resp)
	if err != nil {
		return 0, err
	}
	if resp.Value == 0 {
		return 0, fmt.Errorf("Gandi DNS: Could not create new zone version")
	}
	return resp.Value, nil
}

func (d *DNSProvider) addTXTRecord(zoneID int, version int, name string, value string, ttl int) error {
	resp := &responseStruct{}
	err := rpcCall(&methodCall{
		MethodName: "domain.zone.record.add",
		Params: []param{
			paramString{Value: d.apiKey},
			paramInt{Value: zoneID},
			paramInt{Value: version},
			paramStruct{
				StructMembers: []structMember{
					structMemberString{
						Name:  "type",
						Value: "TXT",
					}, structMemberString{
						Name:  "name",
						Value: name,
					}, structMemberString{
						Name:  "value",
						Value: value,
					}, structMemberInt{
						Name:  "ttl",
						Value: ttl,
					}},
			},
		},
	}, resp)
	if err != nil {
		return err
	}
	return nil
}

func (d *DNSProvider) setZoneVersion(zoneID int, version int) error {
	resp := &responseBool{}
	err := rpcCall(&methodCall{
		MethodName: "domain.zone.version.set",
		Params: []param{
			paramString{Value: d.apiKey},
			paramInt{Value: zoneID},
			paramInt{Value: version},
		},
	}, resp)
	if err != nil {
		return err
	}
	if !resp.Value {
		return fmt.Errorf("Gandi DNS: could not set zone version")
	}
	return nil
}

func (d *DNSProvider) setZone(domain string, zoneID int) error {
	resp := &responseStruct{}
	err := rpcCall(&methodCall{
		MethodName: "domain.zone.set",
		Params: []param{
			paramString{Value: d.apiKey},
			paramString{Value: domain},
			paramInt{Value: zoneID},
		},
	}, resp)
	if err != nil {
		return err
	}
	var respZoneID int
	for _, member := range resp.StructMembers {
		if member.Name == "zone_id" {
			respZoneID = member.ValueInt
		}
	}
	if respZoneID != zoneID {
		return fmt.Errorf("Gandi DNS: Could not set new zone_id")
	}
	return nil
}

func (d *DNSProvider) deleteZone(zoneID int) error {
	resp := &responseBool{}
	err := rpcCall(&methodCall{
		MethodName: "domain.zone.delete",
		Params: []param{
			paramString{Value: d.apiKey},
			paramInt{Value: zoneID},
		},
	}, resp)
	if err != nil {
		return err
	}
	if !resp.Value {
		return fmt.Errorf("Gandi DNS: could not delete zone_id")
	}
	return nil
}
