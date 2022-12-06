package internal

import (
	"fmt"
	"strings"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

type Wrapper struct {
	session *session.Session
}

func NewWrapper(sess *session.Session) *Wrapper {
	return &Wrapper{session: sess}
}

func (w Wrapper) AddTXTRecord(fqdn, domain, value string, ttl int) error {
	service := services.GetDnsDomainService(w.session)

	domainID, err := getDomainID(service, domain)
	if err != nil {
		return fmt.Errorf("failed to get domain ID: %w", err)
	}

	service.Options.Id = domainID

	if _, err := service.CreateTxtRecord(sl.String(fqdn), sl.String(value), sl.Int(ttl)); err != nil {
		return fmt.Errorf("failed to create TXT record: %w", err)
	}

	return nil
}

func (w Wrapper) CleanupTXTRecord(fqdn, domain string) error {
	service := services.GetDnsDomainService(w.session)

	domainID, err := getDomainID(service, domain)
	if err != nil {
		return fmt.Errorf("failed to get domain ID: %w", err)
	}

	service.Options.Id = domainID

	records, err := findTxtRecords(service, fqdn)
	if err != nil {
		return fmt.Errorf("failed to find TXT records: %w", err)
	}

	return deleteResourceRecords(service, records)
}

func getDomainID(service services.Dns_Domain, domain string) (*int, error) {
	res, err := service.GetByDomainName(sl.String(domain))
	if err != nil {
		return nil, err
	}

	for _, r := range res {
		if r.Id == nil || toString(r.Name) != domain {
			continue
		}

		return r.Id, nil
	}

	// The domain was not found by name.
	// For subdomains this is not unusual in softlayer.
	// So in case a subdomain like `sub.toplevel.tld` was used try again using the parent domain
	// (strip the first part in the domain string -> `toplevel.tld`).
	_, parent, found := strings.Cut(domain, ".")
	if !found || !strings.Contains(parent, ".") {
		return nil, fmt.Errorf("no data found for domain: %s", domain)
	}

	return getDomainID(service, parent)
}

func findTxtRecords(service services.Dns_Domain, fqdn string) ([]datatypes.Dns_Domain_ResourceRecord, error) {
	var results []datatypes.Dns_Domain_ResourceRecord

	records, err := service.GetResourceRecords()
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if toString(record.Host) == fqdn && toString(record.Type) == "txt" {
			results = append(results, record)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no data found of fqdn: %s", fqdn)
	}

	return results, nil
}

func deleteResourceRecords(service services.Dns_Domain, records []datatypes.Dns_Domain_ResourceRecord) error {
	resourceRecord := services.GetDnsDomainResourceRecordService(service.Session)

	// TODO maybe a bug: only the last record will be deleted
	for _, record := range records {
		resourceRecord.Options.Id = record.Id
	}

	_, err := resourceRecord.DeleteObject()
	if err != nil {
		return fmt.Errorf("no data found of fqdn: %w", err)
	}

	return nil
}

func toString(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}
