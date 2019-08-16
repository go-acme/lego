package liquidweb

import (
	"log"
	"strconv"

	"github.com/liquidweb/liquidweb-go/network"
)

const defaultBaseURL = "https://api.digitalocean.com"

// txtRecordResponse represents a response from DO's API after making a TXT record
type txtRecordResponse struct {
	DomainRecord record `json:"domain_record"`
}

type record struct {
	ID   int    `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	Data string `json:"data,omitempty"`
	TTL  int    `json:"ttl,omitempty"`
}

type apiError struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// Add LW DNS record.
func (d *DNSProvider) addTxtRecord(domain, fqdn, value string) (*network.DNSRecord, error) {
	dnsRecordParams := &network.DNSRecordParams{
		Name:  fqdn[0 : len(fqdn)-1],
		RData: strconv.Quote(value),
		Type:  "TXT",
		Zone:  d.config.Zone,
	}

	resp, err := d.client.NetworkDNS.Create(dnsRecordParams)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Remove LW DNS record.
func (d *DNSProvider) removeTxtRecord(domain string, recordID int) (bool, error) {
	log.Printf("%v", recordID)
	params := &network.DNSRecordParams{ID: recordID}
	_, err := d.client.NetworkDNS.Delete(params)
	if err != nil {
		return false, err
	}
	return true, nil
}
