package softlayer

import (
	"fmt"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
)

type domainService struct {
	client services.Dns_Domain
}

func (d *domainService) setDomainID(domain string) error {
	res, err := d.client.GetByDomainName(&domain)
	if err != nil {
		return err
	}
	for _, r := range res {
		if r.Name == nil {
			continue
		}
		if *r.Name == domain {
			if r.Id == nil {
				continue
			}
			d.client.Options.Id = r.Id
			return nil
		}
	}
	return fmt.Errorf("no data found of domain: %s", domain)
}

func (d *domainService) findTxtRecords(fqdn string) ([]datatypes.Dns_Domain_ResourceRecord, error) {
	results := []datatypes.Dns_Domain_ResourceRecord{}
	records, err := d.client.GetResourceRecords()
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if record.Host != nil && *record.Host == fqdn {
			if record.Type != nil && *record.Type == "txt" {
				results = append(results, record)
			}
		}
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no data found of fqdn: %s", fqdn)
	}
	return results, nil
}

func (d *domainService) deleteResourceRecords(records []datatypes.Dns_Domain_ResourceRecord) error {
	svc := services.GetDnsDomainResourceRecordService(d.client.Session)
	for _, record := range records {
		svc.Options.Id = record.Id
	}
	_, err := svc.DeleteObject()
	return err
}
