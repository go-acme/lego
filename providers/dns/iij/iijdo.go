package iij

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/iij/doapi"
	"github.com/iij/doapi/protocol"

	"github.com/xenolf/lego/acme"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	AccessKey     string
	SecretKey     string
	DoServiceCode string
}

// DNSProvider implements the acme.ChallengeProvider interface
type DNSProvider struct {
	api    *doapi.API
	config *Config
}

// NewDNSProviderConfig returns a configuration for the DNSProvider
func NewDNSProviderConfig() *Config {
	return &Config{
		AccessKey:     os.Getenv("IIJAPI_ACCESS_KEY"),
		SecretKey:     os.Getenv("IIJAPI_SECRET_KEY"),
		DoServiceCode: os.Getenv("DOSERVICECODE"),
	}
}

// NewDNSProvider returns a DNSProvider instance configured for IIJ DO
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDNSProviderConfig()

	return &DNSProvider{
		api:    doapi.NewAPI(config.AccessKey, config.SecretKey),
		config: config,
	}, nil
}

// ValidateDNSProvider validates DNSProvider Configuration
func (p *DNSProvider) ValidateDNSProvider() bool {
	if len(p.config.AccessKey) > 0 && len(p.config.SecretKey) > 0 && len(p.config.DoServiceCode) > 0 {
		return true
	}

	return false
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (p *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return time.Minute * 2, time.Second * 4
}

// Present creates a TXT record using the specified parameters
func (p *DNSProvider) Present(domain, token, keyAuth string) error {
	_, value, _ := acme.DNS01Record(domain, keyAuth)
	return p.addTxtRecord(domain, value)
}

// CleanUp removes the TXT record matching the specified parameters
func (p *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	_, value, _ := acme.DNS01Record(domain, keyAuth)
	return p.deleteTxtRecord(domain, value)
}

func (p *DNSProvider) addTxtRecord(domain, value string) error {
	owner, zone, err := p.splitDomain(domain)
	if err != nil {
		return err
	}

	request := protocol.RecordAdd{
		DoServiceCode: p.config.DoServiceCode,
		ZoneName:      zone,
		Owner:         owner,
		TTL:           "300",
		RecordType:    "TXT",
		RData:         value,
	}
	response := protocol.RecordAddResponse{}

	if err := doapi.Call(*p.api, request, &response); err != nil {
		return err
	}

	return p.commit()
}

func (p *DNSProvider) deleteTxtRecord(domain, value string) error {
	owner, zone, err := p.splitDomain(domain)
	if err != nil {
		return err
	}

	id, err := p.findTxtRecord(owner, zone, value)
	if err != nil {
		fmt.Println(err)
		return err
	}

	request := protocol.RecordDelete{
		DoServiceCode: p.config.DoServiceCode,
		ZoneName:      zone,
		RecordID:      id,
	}
	response := protocol.RecordDeleteResponse{}

	if err := doapi.Call(*p.api, request, &response); err != nil {
		return err
	}

	return p.commit()
}

func (p *DNSProvider) commit() error {
	request := protocol.Commit{
		DoServiceCode: p.config.DoServiceCode,
	}
	response := protocol.CommitResponse{}

	return doapi.Call(*p.api, request, &response)
}

func (p *DNSProvider) findTxtRecord(owner, zone, value string) (string, error) {
	request := protocol.RecordListGet{
		DoServiceCode: p.config.DoServiceCode,
		ZoneName:      zone,
	}
	response := protocol.RecordListGetResponse{}

	if err := doapi.Call(*p.api, request, &response); err != nil {
		return "", err
	}

	var id string

	for _, record := range response.RecordList {
		if record.Owner == owner && record.RecordType == "TXT" && record.RData == "\""+value+"\"" {
			id = record.Id
		}
	}

	if id == "" {
		return "", fmt.Errorf("%s record in %s not found", owner, zone)
	}

	return id, nil
}

func (p *DNSProvider) listZones() ([]string, error) {
	request := protocol.ZoneListGet{
		DoServiceCode: p.config.DoServiceCode,
	}
	response := protocol.ZoneListGetResponse{}

	if err := doapi.Call(*p.api, request, &response); err != nil {
		return nil, err
	}

	return response.ZoneList, nil
}

func (p *DNSProvider) splitDomain(domain string) (string, string, error) {
	zones, err := p.listZones()
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(strings.Trim(domain, "."), ".")
	owner := ""
	zone := ""
	found := false
Loop:
	for i := 0; i < len(parts)-1; i++ {
		zone = strings.Join(parts[i:], ".")
		for _, z := range zones {
			if zone == z {
				owner = strings.Join(parts[0:i], ".")
				if owner == "" {
					owner = "_acme-challenge"
				} else {
					owner = "_acme-challenge." + owner
				}
				found = true
				break Loop
			}
		}
	}

	if found {
		return owner, zone, nil
	}

	return "", "", fmt.Errorf("%s not found", domain)
}
